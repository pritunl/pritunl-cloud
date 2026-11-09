package statec

import (
	"context"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/mongo-go-driver/v2/mongo/readconcern"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/sirupsen/logrus"
)

const (
	watchMaxAwait         = 5 * time.Second
	watchNextTimeout      = 30 * time.Second
	watchCloseTimeout     = 10 * time.Second
	watchUnsupportedDelay = 60 * time.Second
)

var (
	ResyncInterval = 5 * time.Minute
	Watch          = &Watcher{
		handlers: map[string][]Handler{},
		filters:  map[string]*keyFilter{},
		scoped:   map[string]*Cache{},
	}
)

type Handler interface {
	Handle(evt *Event)
	Resync()
}

type keyFilter struct {
	only   []string
	ignore []string
	keep   []string
}

func (f *keyFilter) active() bool {
	return f != nil && (len(f.only) > 0 || len(f.ignore) > 0)
}

func keyHasPrefix(key string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if key == prefix || strings.HasPrefix(key, prefix+".") {
			return true
		}
	}
	return false
}

func (f *keyFilter) relevant(keys []string) bool {
	if !f.active() {
		return true
	}

	for _, key := range keys {
		if len(f.only) > 0 && !keyHasPrefix(key, f.only) {
			continue
		}

		if len(f.ignore) > 0 && keyHasPrefix(key, f.ignore) &&
			!keyHasPrefix(key, f.keep) {

			continue
		}

		return true
	}

	return false
}

func keyRegex(keys []string) string {
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, regexp.QuoteMeta(key))
	}
	return "^(" + strings.Join(parts, "|") + ")(\\.|$)"
}

func keyMatch(regex string) bson.M {
	return bson.M{
		"$regexMatch": bson.M{
			"input": "$$k",
			"regex": regex,
		},
	}
}

func keysExpr(changedKeys bson.M, cond interface{}) bson.M {
	return bson.M{
		"$anyElementTrue": bson.A{
			bson.M{
				"$map": bson.M{
					"input": changedKeys,
					"as":    "k",
					"in":    cond,
				},
			},
		},
	}
}

func prefixFilter(filter bson.M, prefix string) bson.M {
	out := bson.M{}

	for key, val := range filter {
		switch key {
		case "$or", "$and", "$nor":
			out[key] = prefixFilters(val, prefix)
		default:
			if strings.HasPrefix(key, "$") {
				out[key] = val
			} else {
				out[prefix+key] = val
			}
		}
	}

	return out
}

func prefixFilters(val interface{}, prefix string) interface{} {
	out := bson.A{}

	switch items := val.(type) {
	case []bson.M:
		for _, item := range items {
			out = append(out, prefixFilter(item, prefix))
		}
	case bson.A:
		for _, item := range items {
			if m, ok := item.(bson.M); ok {
				out = append(out, prefixFilter(m, prefix))
			} else {
				out = append(out, item)
			}
		}
	case []interface{}:
		for _, item := range items {
			if m, ok := item.(bson.M); ok {
				out = append(out, prefixFilter(m, prefix))
			} else {
				out = append(out, item)
			}
		}
	default:
		return val
	}

	return out
}

func (f *keyFilter) expr() interface{} {
	conds := []bson.M{}

	if len(f.only) > 0 {
		conds = append(conds, keyMatch(keyRegex(f.only)))
	}

	if len(f.ignore) > 0 {
		notIgnored := bson.M{
			"$not": bson.A{keyMatch(keyRegex(f.ignore))},
		}

		if len(f.keep) > 0 {
			conds = append(conds, bson.M{
				"$or": bson.A{
					notIgnored,
					keyMatch(keyRegex(f.keep)),
				},
			})
		} else {
			conds = append(conds, notIgnored)
		}
	}

	if len(conds) == 1 {
		return conds[0]
	}

	return bson.M{
		"$and": conds,
	}
}

type Watcher struct {
	lock            sync.Mutex
	handlers        map[string][]Handler
	filters         map[string]*keyFilter
	scoped          map[string]*Cache
	started         bool
	healthy         bool
	generation      uint64
	pipelineVersion uint64
	openedVersion   uint64
	events          uint64
	lastEvent       time.Time
	lastResync      time.Time
}

func (w *Watcher) Register(coll string, hand Handler) {
	w.lock.Lock()
	w.handlers[coll] = append(w.handlers[coll], hand)
	w.lock.Unlock()
}

func (w *Watcher) scopeIds(c *Cache) {
	w.lock.Lock()
	w.scoped[c.coll] = c
	w.lock.Unlock()
}

func (w *Watcher) bump(coll, reason string) {
	w.lock.Lock()
	w.pipelineVersion += 1
	version := w.pipelineVersion
	w.lock.Unlock()

	if settings.System.StateDebug {
		logrus.WithFields(logrus.Fields{
			"cache":   coll,
			"reason":  reason,
			"version": int(version),
		}).Info("statec: Change stream pipeline changed")
	}
}

func (w *Watcher) pipelineDirty() bool {
	w.lock.Lock()
	dirty := w.pipelineVersion != w.openedVersion
	w.lock.Unlock()
	return dirty
}

func (w *Watcher) resyncScoped() {
	w.lock.Lock()
	caches := make([]*Cache, 0, len(w.scoped))
	for _, c := range w.scoped {
		caches = append(caches, c)
	}
	w.lock.Unlock()

	for _, c := range caches {
		c.Resync()
	}
}

func (w *Watcher) filter(coll string) (filter *keyFilter) {
	w.lock.Lock()
	filter = w.filters[coll]
	if filter == nil {
		filter = &keyFilter{}
		w.filters[coll] = filter
	}
	w.lock.Unlock()
	return
}

func (w *Watcher) Only(coll string, keys ...string) {
	filter := w.filter(coll)
	filter.only = append(filter.only, keys...)
}

func (w *Watcher) Ignore(coll string, keys ...string) {
	filter := w.filter(coll)
	filter.ignore = append(filter.ignore, keys...)
}

func (w *Watcher) Keep(coll string, keys ...string) {
	filter := w.filter(coll)
	filter.keep = append(filter.keep, keys...)
}

func (w *Watcher) Status() (healthy bool, generation uint64) {
	w.lock.Lock()
	healthy = w.healthy
	generation = w.generation
	w.lock.Unlock()
	return
}

func (w *Watcher) Stats() (events uint64, lastEvent time.Time) {
	w.lock.Lock()
	events = w.events
	lastEvent = w.lastEvent
	w.lock.Unlock()
	return
}

func (w *Watcher) Ensure() {
	w.lock.Lock()
	if w.started {
		w.lock.Unlock()
		return
	}
	w.started = true
	w.lock.Unlock()

	go w.run()
}

func (w *Watcher) collections() (plain []string, filtered []string,
	scoped []*Cache) {

	plain = []string{}
	filtered = []string{}
	scoped = []*Cache{}

	w.lock.Lock()
	for coll, hands := range w.handlers {
		c := w.scoped[coll]
		if c != nil && len(hands) == 1 {
			scoped = append(scoped, c)
		} else if w.filters[coll].active() {
			filtered = append(filtered, coll)
		} else {
			plain = append(plain, coll)
		}
	}
	w.lock.Unlock()

	sort.Strings(plain)
	sort.Strings(filtered)
	sort.Slice(scoped, func(i, j int) bool {
		return scoped[i].coll < scoped[j].coll
	})

	return
}

func (w *Watcher) scopedClause(c *Cache, changedKeys bson.M) bson.M {
	ids := c.Ids()
	scope := c.scope()
	filter := w.filter(c.coll)

	known := bson.M{
		"documentKey._id": bson.M{
			"$in": ids,
		},
	}
	if filter.active() {
		known["$or"] = bson.A{
			bson.M{
				"operationType": bson.M{
					"$ne": opUpdate,
				},
			},
			bson.M{
				"$expr": keysExpr(changedKeys, filter.expr()),
			},
		}
	}

	clauses := bson.A{
		bson.M{
			"operationType": bson.M{
				"$in": bson.A{opDrop, opRename},
			},
		},
		known,
	}

	if scope.Ok {
		insert := bson.M{
			"operationType": bson.M{
				"$in": bson.A{opInsert, opReplace},
			},
		}
		for key, val := range prefixFilter(scope.Filter, "fullDocument.") {
			insert[key] = val
		}
		clauses = append(clauses, insert)
	}

	if len(c.keys) > 0 {
		clauses = append(clauses, bson.M{
			"operationType": opUpdate,
			"$expr": keysExpr(changedKeys,
				keyMatch(keyRegex(c.keys))),
		})
	}

	return bson.M{
		"ns.coll": c.coll,
		"$or":     clauses,
	}
}

func (w *Watcher) pipeline() mongo.Pipeline {
	changedKeys := bson.M{
		"$concatArrays": bson.A{
			bson.M{
				"$map": bson.M{
					"input": bson.M{
						"$objectToArray": bson.M{
							"$ifNull": bson.A{
								"$updateDescription.updatedFields",
								bson.M{},
							},
						},
					},
					"as": "f",
					"in": "$$f.k",
				},
			},
			bson.M{
				"$ifNull": bson.A{
					"$updateDescription.removedFields",
					bson.A{},
				},
			},
			bson.M{
				"$map": bson.M{
					"input": bson.M{
						"$ifNull": bson.A{
							"$updateDescription.truncatedArrays",
							bson.A{},
						},
					},
					"as": "t",
					"in": "$$t.field",
				},
			},
		},
	}

	plain, filtered, scoped := w.collections()

	clauses := bson.A{
		bson.M{
			"ns.coll": bson.M{
				"$in": plain,
			},
		},
		bson.M{
			"operationType": bson.M{
				"$in": bson.A{opInvalidate, opDropDb},
			},
		},
	}

	for _, coll := range filtered {
		clauses = append(clauses, bson.M{
			"ns.coll": coll,
			"$or": bson.A{
				bson.M{
					"operationType": bson.M{
						"$ne": opUpdate,
					},
				},
				bson.M{
					"$expr": keysExpr(changedKeys, w.filter(coll).expr()),
				},
			},
		})
	}

	for _, c := range scoped {
		clauses = append(clauses, w.scopedClause(c, changedKeys))
	}

	match := bson.M{
		"$or": clauses,
	}

	project := bson.M{
		"_id":           1,
		"operationType": 1,
		"ns":            1,
		"documentKey":   1,
		"clusterTime":   1,
		"doc": bson.M{
			"node":       "$fullDocument.node",
			"pool":       "$fullDocument.pool",
			"datacenter": "$fullDocument.datacenter",
			"zone":       "$fullDocument.zone",
			"vpc":        "$fullDocument.vpc",
			"type":       "$fullDocument.type",
			"pod":        "$fullDocument.pod",
			"unit":       "$fullDocument.unit",
			"domain":     "$fullDocument.domain",
		},
		"updated_keys": bson.M{
			"$map": bson.M{
				"input": bson.M{
					"$objectToArray": "$updateDescription.updatedFields",
				},
				"as": "f",
				"in": "$$f.k",
			},
		},
		"removed_fields":   "$updateDescription.removedFields",
		"truncated_arrays": "$updateDescription.truncatedArrays.field",
	}

	return mongo.Pipeline{
		{{"$match", match}},
		{{"$project", project}},
	}
}

func (w *Watcher) dispatch(evt *Event) {
	w.lock.Lock()
	w.events += 1
	w.lastEvent = time.Now()
	hands := w.handlers[evt.Ns.Coll]
	w.lock.Unlock()

	id, _ := evt.DocId()

	if settings.System.StateDebug {
		logrus.WithFields(logrus.Fields{
			"collection": evt.Ns.Coll,
			"operation":  evt.OperationType,
			"id":         id.Hex(),
			"handlers":   len(hands),
			"updated":    debugKeys(evt.UpdatedKeys),
			"removed":    debugKeys(evt.RemovedFields),
			"truncated":  debugKeys(evt.TruncatedArrays),
		}).Info("statec: Change event")
	}

	for _, hand := range hands {
		hand.Handle(evt)
	}
}

func (w *Watcher) resyncAll(reason string) {
	w.lock.Lock()
	hands := []Handler{}
	for _, collHands := range w.handlers {
		hands = append(hands, collHands...)
	}
	w.lock.Unlock()

	if settings.System.StateDebug {
		logrus.WithFields(logrus.Fields{
			"reason":   reason,
			"handlers": len(hands),
		}).Info("statec: Resync all")
	}

	for _, hand := range hands {
		hand.Resync()
	}
}

func (w *Watcher) open(token bson.Raw, reason string) (
	stream *mongo.ChangeStream, err error) {

	defer func() {
		if err != nil {
			w.lock.Lock()
			w.healthy = false
			w.lock.Unlock()
		}
	}()

	db := database.GetDatabase()
	if db == nil {
		err = &errortypes.DatabaseError{
			errors.New("statec: Database not connected"),
		}
		return
	}
	defer db.Close()

	opts := options.ChangeStream().
		SetMaxAwaitTime(watchMaxAwait)
	if token != nil {
		opts.SetResumeAfter(token)
	}

	mdb := db.Nodes().Database()
	watchDb := mdb.Client().Database(
		mdb.Name(),
		options.Database().SetReadConcern(readconcern.Majority()),
	)

	w.lock.Lock()
	version := w.pipelineVersion
	w.lock.Unlock()

	stream, err = watchDb.Watch(
		context.Background(),
		w.pipeline(),
		opts,
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	w.lock.Lock()
	w.healthy = true
	if token == nil {
		w.generation += 1
	}
	generation := w.generation
	w.openedVersion = version
	w.lastResync = time.Now()
	w.lock.Unlock()

	w.resyncScoped()

	logrus.WithFields(logrus.Fields{
		"generation": int(generation),
		"resumed":    token != nil,
		"reason":     reason,
	}).Info("statec: Change stream opened")

	return
}

func (w *Watcher) consume(stream *mongo.ChangeStream) (
	token bson.Raw, reopen bool, invalidated bool, err error) {

	defer func() {
		panc := recover()
		if panc != nil {
			logrus.WithFields(logrus.Fields{
				"trace": string(debug.Stack()),
				"panic": panc,
			}).Error("statec: Panic in change stream")

			err = &errortypes.UnknownError{
				errors.Newf("statec: Change stream panic %v", panc),
			}
			reopen = false
		}

		if !reopen {
			w.lock.Lock()
			w.healthy = false
			w.lock.Unlock()
		}

		ctx, cancel := context.WithTimeout(
			context.Background(), watchCloseTimeout)
		_ = stream.Close(ctx)
		cancel()
	}()

	if cur := stream.ResumeToken(); cur != nil {
		token = bson.Raw(append([]byte(nil), cur...))
	}

	for {
		if constants.Shutdown {
			return
		}

		if w.pipelineDirty() {
			reopen = true
			return
		}

		w.lock.Lock()
		resync := time.Since(w.lastResync) > ResyncInterval
		if resync {
			w.lastResync = time.Now()
		}
		w.lock.Unlock()

		if resync {
			w.resyncAll("interval")
		}

		ctx, cancel := context.WithTimeout(
			context.Background(), watchNextTimeout)
		ok := stream.TryNext(ctx)
		cancel()

		if ok {
			evt := &Event{}
			e := bson.Unmarshal(stream.Current, evt)
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"error": e,
				}).Error("statec: Failed to decode change event")

				w.resyncAll("decode_error")
			} else {
				switch evt.OperationType {
				case opInvalidate:
					w.resyncAll("invalidate")
					invalidated = true
					return
				case opDropDb:
					w.resyncAll("drop_database")
				default:
					w.dispatch(evt)
				}
			}

			if cur := stream.ResumeToken(); cur != nil {
				token = bson.Raw(append([]byte(nil), cur...))
			}

			continue
		}

		e := stream.Err()
		if e != nil {
			err = database.ParseError(e)
			return
		}

		if stream.ID() == 0 {
			err = &errortypes.DatabaseError{
				errors.New("statec: Change stream closed by server"),
			}
			return
		}

		if cur := stream.ResumeToken(); cur != nil {
			token = bson.Raw(append([]byte(nil), cur...))
		}
	}
}

func (w *Watcher) run() {
	var token bson.Raw
	reason := "start"
	unsupported := false

	for {
		if constants.Shutdown {
			return
		}

		if token == nil {
			reason = "start"
		}

		stream, err := w.open(token, reason)
		if err != nil {
			if _, ok := err.(*database.UnsupportedError); ok {
				if !unsupported {
					logrus.WithFields(logrus.Fields{
						"error": err,
					}).Error("statec: Change streams unavailable, " +
						"polling state")
					unsupported = true
				}

				time.Sleep(watchUnsupportedDelay)
				continue
			}

			if _, ok := err.(*database.ResumeError); ok && token != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Warn("statec: Change stream resume failed, resetting")

				token = nil
				continue
			}

			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("statec: Failed to open change stream")

			time.Sleep(constants.RetryDelay)
			continue
		}
		unsupported = false

		newToken, reopen, invalidated, err := w.consume(stream)
		if newToken != nil {
			token = newToken
		}

		if constants.Shutdown {
			return
		}

		if invalidated {
			logrus.Warn("statec: Change stream invalidated, resetting")
			token = nil
			continue
		}

		if reopen {
			reason = "pipeline"
			continue
		}

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("statec: Change stream failed")

			if _, ok := err.(*database.ResumeError); ok {
				token = nil
			}
			reason = "resume"

			time.Sleep(constants.RetryDelay)
		}
	}
}
