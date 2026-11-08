package statec

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/sirupsen/logrus"
)

type Scope struct {
	Filter bson.M
	Key    string
	Ok     bool
}

type Cache struct {
	coll       string
	keys       []string
	scope      func() Scope
	match      func(evt *Event) bool
	filter     *keyFilter
	projection interface{}
	scoped     bool

	lock       sync.Mutex
	docs       map[bson.ObjectID]bson.Raw
	touched    map[bson.ObjectID]struct{}
	loading    bool
	loaded     bool
	resync     bool
	scopeKey   string
	generation uint64
	version    uint64
}

func NewCache(coll string, keys []string, scope func() Scope,
	match func(evt *Event) bool) (c *Cache) {

	c = &Cache{
		coll:    coll,
		keys:    keys,
		scope:   scope,
		match:   match,
		filter:  Watch.filter(coll),
		docs:    map[bson.ObjectID]bson.Raw{},
		touched: map[bson.ObjectID]struct{}{},
	}

	Watch.Register(coll, c)

	return
}

func (c *Cache) SetProjection(projection interface{}) *Cache {
	c.projection = projection
	return c
}

func (c *Cache) Only(keys ...string) *Cache {
	Watch.Only(c.coll, keys...)
	return c
}

func (c *Cache) Ignore(keys ...string) *Cache {
	Watch.Ignore(c.coll, keys...)
	return c
}

func (c *Cache) Keep(keys ...string) *Cache {
	Watch.Keep(c.coll, keys...)
	return c
}

func (c *Cache) ScopeIds() *Cache {
	c.scoped = true
	Watch.scopeIds(c)
	return c
}

func (c *Cache) Collection() string {
	return c.coll
}

func (c *Cache) Ids() (ids []bson.ObjectID) {
	c.lock.Lock()
	ids = make([]bson.ObjectID, 0, len(c.docs))
	for id := range c.docs {
		ids = append(ids, id)
	}
	c.lock.Unlock()

	sort.Slice(ids, func(i, j int) bool {
		return bytes.Compare(ids[i][:], ids[j][:]) < 0
	})

	return
}

func (c *Cache) changed(reason string) {
	if c.scoped {
		Watch.bump(c.coll, reason)
	}
}

func sameIds(a, b map[bson.ObjectID]bson.Raw) bool {
	if len(a) != len(b) {
		return false
	}

	for id := range a {
		if _, ok := b[id]; !ok {
			return false
		}
	}

	return true
}

func (c *Cache) Version() uint64 {
	c.lock.Lock()
	version := c.version
	c.lock.Unlock()
	return version
}

func (c *Cache) Len() int {
	c.lock.Lock()
	n := len(c.docs)
	c.lock.Unlock()
	return n
}

func (c *Cache) Resync() {
	c.lock.Lock()
	c.resync = true
	c.lock.Unlock()
}

func (c *Cache) debugResync(evt *Event, reason string) {
	if !settings.System.StateDebug {
		return
	}

	logrus.WithFields(logrus.Fields{
		"cache":     c.coll,
		"operation": evt.OperationType,
		"reason":    reason,
	}).Info("statec: Cache resync")
}

func (c *Cache) Handle(evt *Event) {
	if evt.IsStructural() {
		c.debugResync(evt, "structural")
		c.Resync()
		return
	}

	id, ok := evt.DocId()
	if !ok {
		c.debugResync(evt, "invalid_id")
		c.Resync()
		return
	}

	c.lock.Lock()

	reason := ""
	if c.loading {
		reason = "loading"
	} else {
		_, has := c.docs[id]

		switch evt.OperationType {
		case opInsert, opReplace:
			if has {
				reason = "cached"
			} else if c.match == nil {
				reason = "unscoped"
			} else if c.match(evt) {
				reason = "match"
			}
		case opUpdate:
			if !c.filter.relevant(evt.ChangedKeys()) {
				break
			}

			if has {
				reason = "cached"
			} else if len(c.keys) > 0 && evt.Touches(c.keys...) {
				reason = "scope_key"
			}
		case opDelete:
			if has {
				reason = "cached"
			}
		}
	}

	if reason != "" {
		c.touched[id] = struct{}{}
	}
	pending := len(c.touched)

	c.lock.Unlock()

	if reason != "" && settings.System.StateDebug {
		logrus.WithFields(logrus.Fields{
			"cache":     c.coll,
			"operation": evt.OperationType,
			"id":        id.Hex(),
			"reason":    reason,
			"pending":   pending,
		}).Info("statec: Cache touched")
	}
}

func (c *Cache) restore(touched map[bson.ObjectID]struct{}) {
	c.lock.Lock()
	for id := range touched {
		c.touched[id] = struct{}{}
	}
	c.loading = false
	c.lock.Unlock()
}

func (c *Cache) find(db *database.Database, filter bson.M) (
	docs map[bson.ObjectID]bson.Raw, err error) {

	coll := db.GetCollection(c.coll)
	docs = map[bson.ObjectID]bson.Raw{}

	opts := options.Find()
	if c.projection != nil {
		opts.SetProjection(c.projection)
	}

	cursor, err := coll.Find(db, filter, opts)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		raw := cursor.Current

		id, ok := raw.Lookup("_id").ObjectIDOK()
		if !ok {
			continue
		}

		docs[id] = bson.Raw(append([]byte(nil), raw...))
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (c *Cache) Sync(db *database.Database) (err error) {
	healthy, generation := Watch.Status()

	c.lock.Lock()
	scope := c.scope()
	reason := ""
	switch {
	case !c.loaded:
		reason = "initial"
	case c.resync:
		reason = "resync"
	case !healthy:
		reason = "unhealthy"
	case generation != c.generation:
		reason = "generation"
	case scope.Key != c.scopeKey:
		reason = "scope"
	}
	full := reason != ""
	touched := c.touched
	c.touched = map[bson.ObjectID]struct{}{}
	if full && scope.Ok {
		c.loading = true
	}
	c.lock.Unlock()

	if !scope.Ok {
		c.lock.Lock()
		cleared := len(c.docs)
		if cleared > 0 {
			c.version += 1
		}
		c.docs = map[bson.ObjectID]bson.Raw{}
		c.loaded = true
		c.resync = false
		scopeChanged := scope.Key != c.scopeKey
		c.scopeKey = scope.Key
		c.generation = generation
		c.lock.Unlock()

		if cleared > 0 || scopeChanged {
			c.changed("disabled")
		}

		if full && settings.System.StateDebug {
			logrus.WithFields(logrus.Fields{
				"cache":   c.coll,
				"reason":  reason,
				"cleared": cleared,
			}).Info("statec: Cache disabled")
		}
		return
	}

	if full {
		start := time.Now()

		docs, e := c.find(db, scope.Filter)
		if e != nil {
			c.restore(touched)
			err = e
			return
		}

		c.lock.Lock()
		idsChanged := scope.Key != c.scopeKey || !sameIds(c.docs, docs)
		c.docs = docs
		c.loading = false
		c.loaded = true
		c.resync = false
		c.scopeKey = scope.Key
		c.generation = generation
		c.version += 1
		c.lock.Unlock()

		if idsChanged {
			c.changed(reason)
		}

		if settings.System.StateDebug {
			logrus.WithFields(logrus.Fields{
				"cache":      c.coll,
				"reason":     reason,
				"scope":      scope.Key,
				"docs":       len(docs),
				"dropped":    len(touched),
				"generation": int(generation),
				"duration":   fmt.Sprintf("%v", time.Since(start)),
			}).Info("statec: Cache loaded")
		}
		return
	}

	if len(touched) == 0 {
		return
	}

	start := time.Now()

	ids := make([]bson.ObjectID, 0, len(touched))
	for id := range touched {
		ids = append(ids, id)
	}

	filter := bson.M{
		"$and": []bson.M{
			scope.Filter,
			{
				"_id": bson.M{
					"$in": ids,
				},
			},
		},
	}

	docs, e := c.find(db, filter)
	if e != nil {
		c.restore(touched)
		err = e
		return
	}

	added := 0
	updated := 0
	removed := 0

	c.lock.Lock()
	for id := range touched {
		doc, ok := docs[id]
		cur, has := c.docs[id]
		if ok {
			if !has {
				c.docs[id] = doc
				c.version += 1
				added += 1
			} else if !bytes.Equal(cur, doc) {
				c.docs[id] = doc
				c.version += 1
				updated += 1
			}
		} else if has {
			delete(c.docs, id)
			c.version += 1
			removed += 1
		}
	}
	size := len(c.docs)
	c.lock.Unlock()

	if added > 0 || removed > 0 {
		c.changed("refresh")
	}

	if settings.System.StateDebug {
		logrus.WithFields(logrus.Fields{
			"cache":    c.coll,
			"touched":  len(touched),
			"added":    added,
			"updated":  updated,
			"removed":  removed,
			"docs":     size,
			"duration": fmt.Sprintf("%v", time.Since(start)),
		}).Info("statec: Cache refreshed")
	}

	return
}

func (c *Cache) Raws() (raws []bson.Raw) {
	c.lock.Lock()
	ids := make([]bson.ObjectID, 0, len(c.docs))
	for id := range c.docs {
		ids = append(ids, id)
	}

	sort.Slice(ids, func(i, j int) bool {
		return bytes.Compare(ids[i][:], ids[j][:]) < 0
	})

	raws = make([]bson.Raw, 0, len(ids))
	for _, id := range ids {
		raws = append(raws, c.docs[id])
	}
	c.lock.Unlock()

	return
}

func Decode[T any](c *Cache) (items []*T, err error) {
	raws := c.Raws()
	items = make([]*T, 0, len(raws))

	for _, raw := range raws {
		item := new(T)

		err = bson.Unmarshal(raw, item)
		if err != nil {
			err = &errortypes.ParseError{
				errors.Wrapf(err, "statec: Failed to decode %s document",
					c.coll),
			}
			return
		}

		items = append(items, item)
	}

	return
}
