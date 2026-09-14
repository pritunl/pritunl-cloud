package telemetry

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/psutil"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

const (
	Process = "process"
	Module  = "module"
	Port    = "port"

	componentsTtl   = 24 * time.Hour
	ComponentsLimit = 2048
	componentsName  = 128
)

type Component struct {
	Type string `bson:"type" json:"type"`
	Name string `bson:"name" json:"name"`
}

func (c *Component) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if c.Type != Process && c.Type != Module && c.Type != Port {
		errData = &errortypes.ErrorData{
			Error:   "invalid_type",
			Message: "Invalid component type",
		}
		return
	}

	c.Name = utils.FilterStr(c.Name, componentsName)
	if c.Name == "" {
		errData = &errortypes.ErrorData{
			Error:   "invalid_name",
			Message: "Invalid component name",
		}
		return
	}

	if c.Type == Port {
		proto, portStr, ok := strings.Cut(c.Name, "/")
		if !ok || (proto != "tcp" && proto != "udp") {
			errData = &errortypes.ErrorData{
				Error:   "invalid_name",
				Message: "Invalid port component name",
			}
			return
		}

		port, e := strconv.Atoi(portStr)
		if e != nil || port < 1 || port > 65535 ||
			strconv.Itoa(port) != portStr {

			errData = &errortypes.ErrorData{
				Error:   "invalid_name",
				Message: "Invalid port component name",
			}
			return
		}
	}

	return
}

type componentKey struct {
	Type string
	Name string
}

var (
	componentsLock sync.Mutex
	componentsSeen = map[componentKey]time.Time{}
)

var Components = &Telemetry[[]*Component]{
	TransmitRate: 6 * time.Minute,
	RefreshRate:  1 * time.Minute,
	Relay:        true,
	Refresher:    ComponentsRefresh,
	Validate: func(data []*Component) []*Component {
		if len(data) > ComponentsLimit {
			return data[:ComponentsLimit]
		}
		return data
	},
}

func componentsMark(now time.Time, typ string, names []string) {
	for _, name := range names {
		name = utils.FilterStr(name, componentsName)
		if name == "" {
			continue
		}

		componentsSeen[componentKey{
			Type: typ,
			Name: name,
		}] = now
	}
}

func ComponentsRefresh() (components []*Component, err error) {
	if Mode == Namespace {
		return
	}

	now := time.Now()

	procs, e := psutil.GetProcessNames()
	if e != nil {
		logrus.WithFields(logrus.Fields{
			"error": e,
		}).Error("telemetry: Failed to get process names")
	}

	mods, e := psutil.GetModuleNames()
	if e != nil {
		logrus.WithFields(logrus.Fields{
			"error": e,
		}).Error("telemetry: Failed to get module names")
	}

	ports := []string{}
	listeners, e := psutil.GetListeners()
	if e != nil {
		logrus.WithFields(logrus.Fields{
			"error": e,
		}).Error("telemetry: Failed to get listeners")
	} else {
		portsSeen := map[string]bool{}
		for _, listener := range listeners {
			port := fmt.Sprintf("%s/%d", listener.Protocol, listener.Port)
			if portsSeen[port] {
				continue
			}

			portsSeen[port] = true
			ports = append(ports, port)
		}
	}

	componentsLock.Lock()
	defer componentsLock.Unlock()

	componentsMark(now, Process, procs)
	componentsMark(now, Module, mods)
	componentsMark(now, Port, ports)

	components = []*Component{}
	for key, seen := range componentsSeen {
		if now.Sub(seen) > componentsTtl {
			delete(componentsSeen, key)
			continue
		}

		components = append(components, &Component{
			Type: key.Type,
			Name: key.Name,
		})
	}

	sort.Slice(components, func(i, j int) bool {
		if components[i].Type != components[j].Type {
			return components[i].Type < components[j].Type
		}
		return components[i].Name < components[j].Name
	})

	return
}

func GetUpdates() (updates []*Update, components []*Component) {
	updates, ok := Updates.Get()
	if !ok {
		updates = nil
	}

	components, ok = Components.Get()
	if !ok {
		components = nil
	}

	if updates != nil && components == nil {
		components = Components.Current()
		if components == nil {
			components = []*Component{}
		}
	}

	return
}

func init() {
	Register(Components)
}
