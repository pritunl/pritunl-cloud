package iptables

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
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

func loadIptables(namespace string, state *State, ipv6 bool) (err error) {
	Lock()
	defer Unlock()

	iptablesCmd := getIptablesCmd(ipv6)

	output := ""
	if namespace == "0" {
		output, err = utils.ExecOutput("", iptablesCmd, "-S")
		if err != nil {
			return
		}
	} else {
		output, err = utils.ExecOutput("",
			"ip", "netns", "exec", namespace, iptablesCmd, "-S")
		if err != nil {
			return
		}
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
		if namespace != "0" {
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
				if item == "--physdev-out" || item == "-o" {
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

		rules := state.Interfaces[namespace+"-"+iface]
		if rules == nil {
			rules = &Rules{
				Namespace: namespace,
				Interface: iface,
				Ingress:   [][]string{},
				Ingress6:  [][]string{},
				Holds:     [][]string{},
				Holds6:    [][]string{},
			}
			state.Interfaces[namespace+"-"+iface] = rules
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
		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "netns",
			"add", rules.Namespace,
		)
		if err != nil {
			return
		}

		oldRules := oldState.Interfaces[rules.Namespace+"-"+rules.Interface]
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

	lockId := stateLock.Lock()
	defer stateLock.Unlock(lockId)

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

		newState.Interfaces["0-host"] = generate("0", "host", ingress)
	}

	for _, inst := range instances {
		for i := range inst.Virt.NetworkAdapters {
			namespace := vm.GetNamespace(inst.Id, i)
			iface := vm.GetIface(inst.Id, i)

			_, ok := newState.Interfaces[namespace+"-"+iface]
			if ok {
				logrus.WithFields(logrus.Fields{
					"namespace": namespace,
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

			rules := generate(namespace, "br0", ingress)
			newState.Interfaces[namespace+"-"+"br0"] = rules
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

	if !node.Self.Firewall {
		return
	}

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
		Lock()
		output, e := utils.ExecCombinedOutput("", "iptables", cmd...)
		Unlock()
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
		Lock()
		output, e := utils.ExecCombinedOutput("", "ip6tables", cmd...)
		Unlock()
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

	_, err = utils.ExecCombinedOutputLogged(
		nil, "sysctl", "-w", "net.ipv6.conf.all.accept_ra=2",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil, "sysctl", "-w", "net.ipv6.conf.default.accept_ra=2",
	)
	if err != nil {
		return
	}

	output, err := utils.ExecOutputLogged(
		nil, "ip", "-o", "link", "show",
	)
	if err != nil {
		return
	}

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || len(fields[1]) < 2 {
			continue
		}
		iface := strings.Split(fields[1][:len(fields[1])-1], "@")[0]

		utils.ExecCombinedOutput("",
			"sysctl", "-w",
			fmt.Sprintf("net.ipv6.conf.%s.accept_ra=2", iface),
		)
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil, "sysctl", "-w", "net.ipv4.ip_forward=1",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil, "sysctl", "-w", "net.ipv6.conf.all.forwarding=1",
	)
	if err != nil {
		return
	}

	state := &State{
		Interfaces: map[string]*Rules{},
	}

	namespaces := []string{"0"}

	output, err = utils.ExecOutputLogged(
		nil,
		"ip", "netns", "list",
	)

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		namespaces = append(
			namespaces,
			fields[0],
		)
	}

	for _, namespace := range namespaces {
		err = loadIptables(namespace, state, false)
		if err != nil {
			return
		}

		err = loadIptables(namespace, state, true)
		if err != nil {
			return
		}
	}

	curState = state

	disks, err := disk.GetNode(db, node.Self.Id)
	if err != nil {
		return
	}

	instances, err := instance.GetAllVirt(db, &bson.M{
		"node": node.Self.Id,
	}, disks)

	err = UpdateState(db, instances)
	if err != nil {
		return
	}

	return
}
