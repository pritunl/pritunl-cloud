package task

import (
	"context"
	"time"

	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/vulnerability"
	"github.com/sirupsen/logrus"
)

const (
	vulnerabilitiesSyncIdle = 30 * time.Second
)

var vulnerabilitiesSync = &Task{
	Name:          "vulnerabilities_sync",
	Type:          Runner,
	Version:       1,
	Duration:      15 * time.Minute,
	Workers:       2,
	RunnerHandler: vulnerabilitiesSyncHandler,
}

func vulnerabilitiesSyncHandler(ctx context.Context,
	db *database.Database) (err error) {

	start := time.Now()
	synced := 0
	failed := 0

	defer func() {
		if synced > 0 || failed > 0 {
			logrus.WithFields(logrus.Fields{
				"synced":   synced,
				"failed":   failed,
				"duration": time.Since(start).String(),
			}).Info("task: Vulnerability sync")
		}
	}()

	err = vulnerability.SyncReset(db, time.Now())
	if err != nil {
		return
	}

	for ctx.Err() == nil {
		vuln, e := vulnerability.SyncClaim(db, time.Now())
		if e != nil {
			err = e
			return
		}

		if vuln == nil {
			err = vulnerability.SyncReset(db, time.Now())
			if err != nil {
				return
			}

			if !Sleep(ctx, vulnerabilitiesSyncIdle) {
				break
			}
			continue
		}

		ok, e := vulnerability.SyncOne(db, vuln)
		if e != nil {
			failed += 1
		}
		if _, notFound := e.(*errortypes.NotFoundError); e != nil &&
			!notFound {

			logrus.WithFields(logrus.Fields{
				"vulnerability": vuln.Id,
				"attempts":      vuln.Attempts + 1,
				"error":         e,
			}).Warn("task: Failed to sync vulnerability")
		}
		if ok {
			synced += 1
		}
	}

	return
}

func init() {
	register(vulnerabilitiesSync)
}
