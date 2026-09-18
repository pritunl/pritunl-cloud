package telemetry

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

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

type ComponentData struct {
	Processes []string `bson:"processes" json:"processes"`
	Modules   []string `bson:"modules" json:"modules"`
	Ports     []string `bson:"ports" json:"ports"`
}

func NewComponentData() *ComponentData {
	return &ComponentData{
		Processes: []string{},
		Modules:   []string{},
		Ports:     []string{},
	}
}

func (c *ComponentData) Len() int {
	if c == nil {
		return 0
	}
	return len(c.Processes) + len(c.Modules) + len(c.Ports)
}

func (c *ComponentData) Normalize() {
	if c.Processes == nil {
		c.Processes = []string{}
	}
	if c.Modules == nil {
		c.Modules = []string{}
	}
	if c.Ports == nil {
		c.Ports = []string{}
	}
}

func validPort(name string) bool {
	proto, portStr, ok := strings.Cut(name, "/")
	if !ok || (proto != "tcp" && proto != "udp") {
		return false
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 ||
		strconv.Itoa(port) != portStr {

		return false
	}

	return true
}

func validateNames(names []string, isPort bool) (
	valid []string, invalid int) {

	valid = []string{}
	seen := map[string]bool{}

	for _, name := range names {
		name = utils.FilterStr(name, componentsName)
		if name == "" || (isPort && !validPort(name)) {
			invalid += 1
			continue
		}

		if seen[name] {
			continue
		}
		seen[name] = true

		valid = append(valid, name)
		if len(valid) >= ComponentsLimit {
			break
		}
	}

	return
}

func (c *ComponentData) Validate() (clean *ComponentData, invalid int) {
	clean = NewComponentData()
	if c == nil {
		return
	}

	var n int
	clean.Processes, n = validateNames(c.Processes, false)
	invalid += n
	clean.Modules, n = validateNames(c.Modules, false)
	invalid += n
	clean.Ports, n = validateNames(c.Ports, true)
	invalid += n

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

var Components = &Telemetry[*ComponentData]{
	TransmitRate: 6 * time.Minute,
	RefreshRate:  1 * time.Minute,
	Relay:        true,
	Refresher:    ComponentsRefresh,
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

func ComponentsRefresh() (components *ComponentData, err error) {
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

	components = NewComponentData()
	for key, seen := range componentsSeen {
		if now.Sub(seen) > componentsTtl {
			delete(componentsSeen, key)
			continue
		}

		switch key.Type {
		case Process:
			components.Processes = append(components.Processes, key.Name)
		case Module:
			components.Modules = append(components.Modules, key.Name)
		case Port:
			components.Ports = append(components.Ports, key.Name)
		}
	}

	sort.Strings(components.Processes)
	sort.Strings(components.Modules)
	sort.Strings(components.Ports)

	return
}

func GetUpdates() (updates []*Update, components *ComponentData) {
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
			components = NewComponentData()
		}
	}

	return
}

func init() {
	Register(Components)
}
