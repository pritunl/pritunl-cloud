package telemetry

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	registry []handler
)

type handler interface {
	getName() string
	Refresh() error
}

type Telemetry[Data any] struct {
	name         string
	Default      Data
	lock         sync.Mutex
	lastTransmit time.Time
	TransmitRate time.Duration
	lastRefresh  time.Time
	RefreshRate  time.Duration
	Refresher    func() (Data, error)
	Validate     func(Data) Data
	data         Data
}

func (r *Telemetry[Data]) getName() string {
	return r.name
}

func (r *Telemetry[Data]) Refresh() (err error) {
	r.lock.Lock()
	lastRefresh := r.lastRefresh
	r.lock.Unlock()
	if time.Since(lastRefresh) < r.RefreshRate {
		return
	}

	data, err := r.Refresher()
	if err != nil {
		return
	}

	r.Set(data)

	return
}

func (r *Telemetry[Data]) Set(data Data) {
	r.lock.Lock()
	r.data = data
	r.lastRefresh = time.Now()
	r.lock.Unlock()
}

func (r *Telemetry[Data]) Get() Data {
	r.lock.Lock()
	lastRefresh := r.lastRefresh
	lastTransmit := r.lastTransmit
	r.lock.Unlock()
	if lastRefresh.IsZero() || time.Since(lastTransmit) < r.TransmitRate {
		var x Data
		return x
	}
	r.lock.Lock()
	r.lastTransmit = time.Now()
	r.lock.Unlock()

	if r.Validate != nil {
		return r.Validate(r.data)
	} else {
		return r.data
	}
}

func Register[Data any](telm *Telemetry[Data]) {
	telm.name = fmt.Sprintf("%T", telm)
	registry = append(registry, telm)
}

func Refresh() {
	for _, telm := range registry {
		err := telm.Refresh()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"kind": telm.getName(),
			}).Error("telemetry: Telemetry refresh failed")
		}
	}
}
