package telemetry

import (
	"fmt"
	"sync"
	"time"

	"github.com/pritunl/pritunl-cloud/utils"
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
	Queue        int
	Refresher    func() (Data, error)
	Validate     func(Data) Data
	data         Data
	dataList     []Data
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

	var data Data
	func() {
		defer utils.RecoverLog("telemetry: Panic in refresh")

		data, err = r.Refresher()
		if err != nil {
			return
		}
	}()

	if r.Queue > 0 {
		if err == nil {
			if r.Validate != nil {
				data = r.Validate(data)
			}
			r.Append(data)
		}

		r.lock.Lock()
		r.lastRefresh = time.Now()
		r.lock.Unlock()
	} else {
		r.Set(data)
	}

	return
}

func (r *Telemetry[Data]) Set(data Data) {
	r.lock.Lock()
	r.data = data
	r.lastRefresh = time.Now()
	r.lock.Unlock()
}

func (r *Telemetry[Data]) Get() (Data, bool) {
	r.lock.Lock()
	lastRefresh := r.lastRefresh
	lastTransmit := r.lastTransmit
	r.lock.Unlock()
	if lastRefresh.IsZero() || time.Since(lastTransmit) < r.TransmitRate {
		var x Data
		return x, false
	}
	r.lock.Lock()
	r.lastTransmit = time.Now()
	r.lock.Unlock()

	if r.Validate != nil {
		return r.Validate(r.data), true
	} else {
		return r.data, true
	}
}

func (r *Telemetry[Data]) Append(items ...Data) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.dataList = append(r.dataList, items...)

	if r.Queue > 0 && len(r.dataList) > r.Queue {
		r.dataList = append([]Data{},
			r.dataList[len(r.dataList)-r.Queue:]...)
	}
}

func (r *Telemetry[Data]) GetAll() (items []Data) {
	r.lock.Lock()
	defer r.lock.Unlock()

	items = r.dataList
	r.dataList = nil
	return
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
