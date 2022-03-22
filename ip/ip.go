package ip

import (
	"encoding/json"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
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
