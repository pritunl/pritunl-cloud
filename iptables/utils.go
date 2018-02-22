package iptables

import (
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

func diffCmd(a, b []string) bool {
	if len(a) != len(b) {
		return true
	}

	for i := range a {
		if a[i] != b[i] {
			return true
		}
	}

	return false
}

func diffRules(a, b *Rules) bool {
	if len(a.Ingress) != len(b.Ingress) ||
		len(a.Ingress6) != len(b.Ingress6) ||
		len(a.Holds) != len(b.Holds) ||
		len(a.Holds6) != len(b.Holds6) {

		return true
	}

	for i := range a.Ingress {
		if diffCmd(a.Ingress[i], b.Ingress[i]) {
			return true
		}
	}
	for i := range a.Ingress6 {
		if diffCmd(a.Ingress6[i], b.Ingress6[i]) {
			return true
		}
	}
	for i := range a.Holds {
		if diffCmd(a.Holds[i], b.Holds[i]) {
			return true
		}
	}
	for i := range a.Holds6 {
		if diffCmd(a.Holds6[i], b.Holds6[i]) {
			return true
		}
	}

	return false
}

func getIptablesCmd(ipv6 bool) string {
	if ipv6 {
		return "ip6tables"
	} else {
		return "iptables"
	}
}

func loadIptables(state *State, ipv6 bool) (err error) {
	iptablesCmd := getIptablesCmd(ipv6)

	output, err := utils.ExecOutput("", iptablesCmd, "-S")
	if err != nil {
		return
	}

	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "pritunl_cloud_rule") &&
			!strings.Contains(line, "pritunl_cloud_hold") {

			continue
		}

		cmd := strings.Fields(line)
		if len(cmd) < 3 {
			logrus.WithFields(logrus.Fields{
				"iptables_rule": line,
			}).Error("iptables: Invalid iptables state")

			err = &errortypes.ParseError{
				errors.New("iptables: Invalid iptables state"),
			}
			return
		}
		cmd = cmd[1:]

		iface := ""
		if strings.Contains(line, "--physdev-out") {
			if cmd[0] != "FORWARD" {
				logrus.WithFields(logrus.Fields{
					"iptables_rule": line,
				}).Error("iptables: Invalid iptables chain")

				err = &errortypes.ParseError{
					errors.New("iptables: Invalid iptables chain"),
				}
				return
			}

			for i, item := range cmd {
				if item == "--physdev-out" {
					if len(cmd) < i+2 {
						logrus.WithFields(logrus.Fields{
							"iptables_rule": line,
						}).Error("iptables: Invalid iptables interface")

						err = &errortypes.ParseError{
							errors.New("iptables: Invalid iptables interface"),
						}
						return
					}
					iface = cmd[i+1]
					break
				}
			}
		} else {
			iface = "host"

			if cmd[0] != "INPUT" {
				logrus.WithFields(logrus.Fields{
					"iptables_rule": line,
				}).Error("iptables: Invalid iptables chain")

				err = &errortypes.ParseError{
					errors.New("iptables: Invalid iptables chain"),
				}
				return
			}
		}

		if iface == "" {
			logrus.WithFields(logrus.Fields{
				"iptables_rule": line,
			}).Error("iptables: Missing iptables interface")

			err = &errortypes.ParseError{
				errors.New("iptables: Missing iptables interface"),
			}
			return
		}

		rules := state.Interfaces[iface]
		if rules == nil {
			rules = &Rules{
				Interface: iface,
				Ingress:   [][]string{},
				Ingress6:  [][]string{},
				Holds:     [][]string{},
				Holds6:    [][]string{},
			}
			state.Interfaces[iface] = rules
		}

		if strings.Contains(line, "pritunl_cloud_hold") {
			if ipv6 {
				rules.Holds6 = append(rules.Holds6, cmd)
			} else {
				rules.Holds = append(rules.Holds, cmd)
			}
		} else {
			if ipv6 {
				rules.Ingress6 = append(rules.Ingress6, cmd)
			} else {
				rules.Ingress = append(rules.Ingress, cmd)
			}
		}
	}

	return
}

func applyState(oldState, newState *State) (err error) {
	oldIfaces := set.NewSet()
	newIfaces := set.NewSet()

	for iface := range oldState.Interfaces {
		oldIfaces.Add(iface)
	}
	for iface := range newState.Interfaces {
		newIfaces.Add(iface)
	}

	oldIfaces.Subtract(newIfaces)
	for iface := range oldIfaces.Iter() {
		err = oldState.Interfaces[iface.(string)].Remove()
		if err != nil {
			return
		}
	}

	for _, rules := range newState.Interfaces {
		oldRules := oldState.Interfaces[rules.Interface]
		if oldRules != nil {
			if !diffRules(oldRules, rules) {
				continue
			}

			if rules.Interface != "host" {
				err = rules.Hold()
				if err != nil {
					return
				}
			}

			err = oldRules.Remove()
			if err != nil {
				return
			}
		}

		err = rules.Apply()
		if err != nil {
			return
		}
	}

	return
}

func UpdateState(db *database.Database, instances []*instance.Instance) (
	err error) {

	stateLock.Lock()
	defer stateLock.Unlock()

	newState := &State{
		Interfaces: map[string]*Rules{},
	}

	if node.Self.Firewall {
		fires, e := firewall.GetRoles(db, node.Self.NetworkRoles)
		if e != nil {
			err = e
			return
		}

		ingress := []*firewall.Rule{}
		for _, fire := range fires {
			ingress = append(ingress, fire.Ingress...)
		}

		rules := generate("host", ingress)
		newState.Interfaces["host"] = rules
	}

	for _, inst := range instances {
		virt := inst.GetVm(nil)

		for i := range virt.NetworkAdapters {
			iface := vm.GetIface(virt.Id, i)

			_, ok := newState.Interfaces[iface]
			if ok {
				logrus.WithFields(logrus.Fields{
					"interface": iface,
				}).Error("iptables: Virtual interface conflict")

				err = &errortypes.ParseError{
					errors.New("iptables: Virtual interface conflict"),
				}
				panic(err)
			}

			fires, e := firewall.GetOrgRoles(db,
				inst.Organization, inst.NetworkRoles)
			if e != nil {
				err = e
				return
			}

			ingress := []*firewall.Rule{}
			for _, fire := range fires {
				ingress = append(ingress, fire.Ingress...)
			}

			rules := generate(iface, ingress)
			newState.Interfaces[iface] = rules
		}
	}

	err = applyState(curState, newState)
	if err != nil {
		return
	}

	curState = newState

	return
}

func Recover() (err error) {
	cmds := [][]string{}

	cmds = append(cmds, []string{
		"-I", "INPUT", "1",
		"-m", "comment",
		"--comment", "pritunl_cloud_rule",
		"-j", "DROP",
	})
	cmds = append(cmds, []string{
		"-I", "INPUT", "1",
		"-m", "conntrack",
		"--ctstate", "INVALID",
		"-m", "comment",
		"--comment", "pritunl_cloud_rule",
		"-j", "DROP",
	})
	cmds = append(cmds, []string{
		"-I", "INPUT", "1",
		"-m", "conntrack",
		"--ctstate", "RELATED,ESTABLISHED",
		"-m", "comment", "--comment", "pritunl_cloud_rule",
		"-j", "ACCEPT",
	})

	for _, cmd := range cmds {
		output, e := utils.ExecCombinedOutput("", "iptables", cmd...)
		if e != nil {
			err = e
			logrus.WithFields(logrus.Fields{
				"command": cmd,
				"output":  output,
				"error":   err,
			}).Error("iptables: Failed to add iptables recover rule")
			return
		}
	}

	for _, cmd := range cmds {
		output, e := utils.ExecCombinedOutput("", "ip6tables", cmd...)
		if e != nil {
			err = e
			logrus.WithFields(logrus.Fields{
				"command": cmd,
				"output":  output,
				"error":   err,
			}).Error("iptables: Failed to add ip6tables recover rule")
			return
		}
	}

	time.Sleep(10 * time.Second)

	err = Init()
	if err != nil {
		return
	}

	return
}

func Init() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	state := &State{
		Interfaces: map[string]*Rules{},
	}

	err = loadIptables(state, false)
	if err != nil {
		return
	}

	err = loadIptables(state, true)
	if err != nil {
		return
	}

	curState = state

	instances, err := instance.GetAll(db, &bson.M{
		"node": node.Self.Id,
	})

	err = UpdateState(db, instances)
	if err != nil {
		return
	}

	return
}
