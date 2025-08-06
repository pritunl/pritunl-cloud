package ip

import (
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	cache     = map[string]map[string]*Iface{}
	cacheTime = map[string]time.Time{}
	cacheLock = sync.Mutex{}
)

type Iface struct {
	Ifindex   int      `json:"ifindex"`
	Ifname    string   `json:"ifname"`
	Flags     []string `json:"flags"`
	Mtu       int      `json:"mtu"`
	Qdisc     string   `json:"qdisc"`
	Operstate string   `json:"operstate"`
	Group     string   `json:"group"`
	Txqlen    int      `json:"txqlen"`
	LinkType  string   `json:"link_type"`
	Address   string   `json:"address"`
	Broadcast string   `json:"broadcast"`
	AddrInfo  []struct {
		Family            string `json:"family"`
		Local             string `json:"local"`
		Prefixlen         int    `json:"prefixlen"`
		Scope             string `json:"scope"`
		Label             string `json:"label,omitempty"`
		ValidLifeTime     int64  `json:"valid_life_time"`
		PreferredLifeTime int64  `json:"preferred_life_time"`
		Broadcast         string `json:"broadcast,omitempty"`
		Dynamic           bool   `json:"dynamic,omitempty"`
		Mngtmpaddr        bool   `json:"mngtmpaddr,omitempty"`
	} `json:"addr_info"`
	Link        string `json:"link,omitempty"`
	Master      string `json:"master,omitempty"`
	LinkIndex   int    `json:"link_index,omitempty"`
	LinkNetnsid int    `json:"link_netnsid,omitempty"`
}

func GetIfaces(namespace string) (ifaces []*Iface, err error) {
	output := ""

	if namespace == "" {
		output, err = utils.ExecOutput(
			"",
			"ip", "-j", "address",
		)
	} else {
		output, err = utils.ExecOutput(
			"",
			"ip", "netns", "exec", namespace,
			"ip", "-j", "address",
		)
	}
	if err != nil {
		return
	}

	ifaces = []*Iface{}

	err = json.Unmarshal([]byte(output), &ifaces)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "ip: Failed to parse ip json output"),
		}
		return
	}

	return
}

func GetIfacesCached(namespace string) (ifacesMap map[string]*Iface, err error) {
	cacheLock.Lock()
	if time.Since(cacheTime[namespace]) < 5*time.Minute {
		ifacesMap = cache[namespace]
		cacheLock.Unlock()
		return
	}
	cacheLock.Unlock()

	ifaces, err := GetIfaces(namespace)
	if err != nil {
		return
	}

	ifacesMap = map[string]*Iface{}
	for _, iface := range ifaces {
		ifacesMap[iface.Ifname] = iface
	}

	cacheLock.Lock()
	cache[namespace] = ifacesMap
	cacheTime[namespace] = time.Now()
	cacheLock.Unlock()

	return
}
