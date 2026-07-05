package iproute

import (
	"encoding/json"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Address struct {
	Family     string `json:"family"`
	Local      string `json:"local"`
	Prefix     int    `json:"prefixlen"`
	Scope      string `json:"scope"`
	Label      string `json:"label"`
	Dynamic    bool   `json:"dynamic"`
	Deprecated bool   `json:"deprecated"`
	Protocol   string `json:"protocol"`
}

func addressRank6(addr *Address) int {
	if !addr.Dynamic {
		return 3
	}
	if addr.Protocol != "kernel_ra" {
		return 2
	}
	return 1
}

type AddressIface struct {
	Name      string     `json:"ifname"`
	State     string     `json:"operstate"`
	Addresses []*Address `json:"addr_info"`
}

func AddressGetIface(namespace, name string) (
	address, address6 *Address, err error) {

	ifaces := []*AddressIface{}

	label := ""
	if strings.Contains(name, ":") {
		label = name
		name = strings.Split(name, ":")[0]
	}

	var output string
	if namespace != "" {
		output, err = utils.ExecOutputLogged(
			[]string{
				"No such file or directory",
				"does not exist",
				"setting the network namespace",
			},
			"ip", "netns", "exec", namespace,
			"ip", "--json",
			"addr", "show",
			"dev", name,
		)
	} else {
		output, err = utils.ExecOutputLogged(
			[]string{
				"No such file or directory",
				"does not exist",
				"setting the network namespace",
			},
			"ip", "--json",
			"addr", "show",
			"dev", name,
		)
	}
	if err != nil {
		return
	}

	if output == "" {
		return
	}

	err = json.Unmarshal([]byte(output), &ifaces)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "iproute: Failed to parse interface address"),
		}
		return
	}

	rank6 := 0
	if label != "" {
		for _, iface := range ifaces {
			if iface.Name == name && iface.Addresses != nil {
				for _, addr := range iface.Addresses {
					if addr.Label == label && addr.Scope == "global" &&
						!addr.Deprecated {

						if address == nil && addr.Family == "inet" {
							address = addr
						} else if addr.Family == "inet6" {
							if addressRank6(addr) > rank6 {
								address6 = addr
								rank6 = addressRank6(addr)
							}
						}
					}
				}
			}
		}
	}

	for _, iface := range ifaces {
		if iface.Name == name && iface.Addresses != nil {
			for _, addr := range iface.Addresses {
				if addr.Scope == "global" && !addr.Deprecated {
					if address == nil && addr.Family == "inet" {
						address = addr
					} else if addr.Family == "inet6" {
						if addressRank6(addr) > rank6 {
							address6 = addr
							rank6 = addressRank6(addr)
						}
					}
				}
			}
		}
	}

	return
}

func AddressGetIfaceMod(namespace, name string) (
	address, address6 *Address, err error) {

	ifaces := []*AddressIface{}

	label := ""
	if strings.Contains(name, ":") {
		label = name
		name = strings.Split(name, ":")[0]
	}

	nameMod := name + "x"
	nameMod6 := name + "y"
	var addressMod *Address
	var address6Mod *Address
	var addressMod6 *Address
	var address6Mod6 *Address
	rankMod6 := 0
	rank6Mod6 := 0

	var output string
	if namespace != "" {
		output, err = utils.ExecOutputLogged(
			[]string{
				"No such file or directory",
				"does not exist",
				"setting the network namespace",
			},
			"ip", "netns", "exec", namespace,
			"ip", "--json",
			"addr", "show",
		)
	} else {
		output, err = utils.ExecOutputLogged(
			[]string{
				"No such file or directory",
				"does not exist",
				"setting the network namespace",
			},
			"ip", "--json",
			"addr", "show",
		)
	}
	if err != nil {
		return
	}

	if output == "" {
		return
	}

	err = json.Unmarshal([]byte(output), &ifaces)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "iproute: Failed to parse interface address"),
		}
		return
	}

	rank6 := 0
	if label != "" {
		for _, iface := range ifaces {
			if strings.HasPrefix(iface.Name, name) && iface.Addresses != nil {
				for _, addr := range iface.Addresses {
					if addr.Label == label && addr.Scope == "global" &&
						!addr.Deprecated {

						if address == nil && addr.Family == "inet" {
							address = addr
						} else if addr.Family == "inet6" {
							if addressRank6(addr) > rank6 {
								address6 = addr
								rank6 = addressRank6(addr)
							}
						}
					}
				}
			}
		}
	}

	for _, iface := range ifaces {
		if iface.Name == name && iface.Addresses != nil {
			for _, addr := range iface.Addresses {
				if addr.Scope == "global" && !addr.Deprecated {
					if address == nil && addr.Family == "inet" {
						address = addr
					} else if addr.Family == "inet6" {
						if addressRank6(addr) > rank6 {
							address6 = addr
							rank6 = addressRank6(addr)
						}
					}
				}
			}
		}

		if iface.Name == nameMod && iface.Addresses != nil {
			for _, addr := range iface.Addresses {
				if addr.Scope == "global" && !addr.Deprecated {
					if addressMod == nil && addr.Family == "inet" {
						addressMod = addr
					} else if addr.Family == "inet6" {
						if addressRank6(addr) > rankMod6 {
							addressMod6 = addr
							rankMod6 = addressRank6(addr)
						}
					}
				}
			}
		}

		if iface.Name == nameMod6 && iface.Addresses != nil {
			for _, addr := range iface.Addresses {
				if addr.Scope == "global" && !addr.Deprecated {
					if address6Mod == nil && addr.Family == "inet" {
						address6Mod = addr
					} else if addr.Family == "inet6" {
						if addressRank6(addr) > rank6Mod6 {
							address6Mod6 = addr
							rank6Mod6 = addressRank6(addr)
						}
					}
				}
			}
		}
	}

	if address6Mod6 != nil {
		address6 = address6Mod6
	} else if addressMod6 != nil {
		address6 = addressMod6
	}

	if addressMod != nil {
		address = addressMod
	} else if address6Mod != nil {
		address = address6Mod
	}

	return
}
