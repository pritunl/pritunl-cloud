package sync

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/statec"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

const (
	stateCompareLogLimit = 10
	stateCompareRecheck  = 5 * time.Second
)

var stateCompare = &stateComparer{}

type stateComparer struct {
	next    time.Time
	recheck bool
	prev    map[string]*statec.Diff
}

func (s *stateComparer) run(stat *state.State, fromStatec bool) {
	if time.Now().Before(s.next) {
		return
	}

	defer func() {
		panc := recover()
		if panc != nil {
			logrus.WithFields(logrus.Fields{
				"trace": string(debug.Stack()),
				"panic": panc,
			}).Error("sync: Panic in state compare")
		}
	}()

	start := time.Now()

	runtimes := &state.Runtimes{}
	runtimes.Init()

	var other *state.State
	var err error
	otherName := "statec"
	if fromStatec {
		otherName = "state"
		other, err = state.GetState(runtimes)
	} else {
		other, err = statec.GetState(runtimes)
	}
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"source": otherName,
			"error":  err,
		}).Error("sync: Failed to get compare state")
		return
	}
	runtimes.Total = time.Since(start)

	start = time.Now()
	var diffs []*statec.Diff
	if fromStatec {
		diffs = statec.Compare(other, stat)
	} else {
		diffs = statec.Compare(stat, other)
	}
	compareDur := time.Since(start)

	cur := map[string]*statec.Diff{}
	persistent := []*statec.Diff{}
	transient := []*statec.Diff{}
	raced := 0
	for _, diff := range diffs {
		ident := diff.Ident()
		cur[ident] = diff

		prev := s.prev[ident]
		if prev == nil {
			transient = append(transient, diff)
		} else if diff.CaughtUp(prev) {
			raced += 1
			transient = append(transient, diff)
		} else {
			persistent = append(persistent, diff)
		}
	}

	recheck := s.recheck
	if !recheck && len(diffs) > 0 {
		s.recheck = true
		s.prev = cur
		s.next = time.Now().Add(stateCompareRecheck)
	} else {
		ttl := time.Duration(settings.System.StateCompareTtl) * time.Second
		if ttl < 0 {
			ttl = 0
		}

		s.recheck = false
		s.prev = nil
		s.next = time.Now().Add(ttl)
	}

	healthy, generation := statec.Watch.Status()
	events, lastEvent := statec.Watch.Stats()

	fields := logrus.Fields{
		"source":            stateSourceName(fromStatec),
		"compare_source":    otherName,
		"compare_build":     fmt.Sprintf("%v", runtimes.Total),
		"compare_runtime":   fmt.Sprintf("%v", compareDur),
		"recheck":           recheck,
		"diffs":             len(diffs),
		"persistent_diffs":  len(persistent),
		"raced_diffs":       raced,
		"stream_healthy":    healthy,
		"stream_generation": int(generation),
		"stream_events":     int(events),
	}
	if !lastEvent.IsZero() {
		fields["stream_last_event"] = fmt.Sprintf("%v",
			time.Since(lastEvent).Round(time.Millisecond))
	}
	for key, dur := range runtimes.State {
		fields[otherName+"_"+utils.ToSnakeCase(key)] = fmt.Sprintf(
			"%v", dur)
	}

	switch {
	case recheck && len(persistent) > 0:
		fields["persistent"] = statec.JoinDiffs(
			persistent, stateCompareLogLimit)
		logrus.WithFields(fields).Warn(
			"sync: State compare persistent mismatch")
	case len(diffs) > 0:
		if settings.System.StateDebug {
			fields["transient"] = statec.JoinDiffs(
				transient, stateCompareLogLimit)
			if recheck {
				logrus.WithFields(fields).Info(
					"sync: State compare transient mismatch")
			} else {
				logrus.WithFields(fields).Info(
					"sync: State compare mismatch, rechecking")
			}
		}
	default:
		logrus.WithFields(fields).Info("sync: State compare consistent")
	}
}

func stateSourceName(fromStatec bool) string {
	if fromStatec {
		return "statec"
	}
	return "state"
}
