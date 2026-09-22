package task

import (
	"context"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/vulnerability"
	"github.com/sirupsen/logrus"
)

const (
	vulnerabilitiesSyncIdle = 30 * time.Second
)

var vulnerabilitiesSync = &Task{
	Name:     "vulnerabilities_sync",
	Type:     Runner,
	Version:  1,
	Duration: 15 * time.Minute,
	Workers: func() int {
		return settings.Telemetry.SyncWorkers
	},
	RunnerHandler: vulnerabilitiesSyncHandler,
}

type vulnerabilitiesSyncState struct {
	synced atomic.Int64
	failed atomic.Int64
	lock   sync.Mutex
	err    error
}

func (s *vulnerabilitiesSyncState) setError(err error) {
	s.lock.Lock()
	if s.err == nil {
		s.err = err
	}
	s.lock.Unlock()
}

func (s *vulnerabilitiesSyncState) getError() (err error) {
	s.lock.Lock()
	err = s.err
	s.lock.Unlock()
	return
}

func vulnerabilitiesSyncHandler(ctx context.Context,
	db *database.Database) (err error) {

	start := time.Now()
	state := &vulnerabilitiesSyncState{}

	defer func() {
		failed := state.failed.Load()
		if failed > 0 {
			synced := state.synced.Load()
			logrus.WithFields(logrus.Fields{
				"synced":   synced,
				"failed":   failed,
				"duration": time.Since(start).Round(time.Millisecond).String(),
			}).Info("task: Vulnerability sync incomplete")
		}
	}()

	err = vulnerability.SyncReset(db, time.Now())
	if err != nil {
		return
	}

	threads := settings.Telemetry.SyncThreads
	if threads < 1 {
		threads = 1
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	waiter := &sync.WaitGroup{}
	for i := 0; i < threads; i++ {
		waiter.Add(1)
		go func(reset bool) {
			defer waiter.Done()

			e := vulnerabilitiesSyncWorker(ctx, state, reset)
			if e != nil {
				state.setError(e)
				cancel()
			}
		}(i == 0)
	}
	waiter.Wait()

	err = state.getError()
	return
}

func vulnerabilitiesSyncWorker(ctx context.Context,
	state *vulnerabilitiesSyncState, reset bool) (err error) {

	defer func() {
		panc := recover()
		if panc != nil {
			logrus.WithFields(logrus.Fields{
				"trace": string(debug.Stack()),
				"panic": panc,
			}).Error("task: Panic in vulnerability sync worker")

			err = &errortypes.UnknownError{
				errors.New("task: Panic in vulnerability sync worker"),
			}
		}
	}()

	db := database.GetDatabase()
	defer db.Close()

	for ctx.Err() == nil {
		err = vulnerabilitiesSyncOne(ctx, db, state, reset)
		if err != nil {
			return
		}
	}

	return
}

func vulnerabilitiesSyncOne(ctx context.Context, db *database.Database,
	state *vulnerabilitiesSyncState, reset bool) (err error) {

	vuln, err := vulnerability.SyncClaim(db, time.Now())
	if err != nil {
		return
	}

	if vuln == nil {
		if reset {
			err = vulnerability.SyncReset(db, time.Now())
			if err != nil {
				return
			}
		}

		Sleep(ctx, vulnerabilitiesSyncIdle)
		return
	}

	synced, err := vulnerability.SyncOne(ctx, db, vuln)
	if err != nil {
		state.failed.Add(1)
		if _, ok := err.(*errortypes.NotFoundError); !ok {
			logrus.WithFields(logrus.Fields{
				"vulnerability": vuln.Id,
				"attempts":      vuln.Attempts + 1,
				"error":         err,
			}).Warn("task: Failed to sync vulnerability")
		}
		err = nil
	}
	if synced {
		state.synced.Add(1)
	}

	return
}

func init() {
	register(vulnerabilitiesSync)
}
