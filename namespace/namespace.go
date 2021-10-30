package namespace

import (
	"path"
	"sync"

	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	lock = sync.Mutex{}
)

func CopyIptables(cur, new string) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		[]string{},
		"ip", "netns", "exec", new,
		"iptables", "-t", "nat", "-F",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{},
		"ip", "netns", "exec", new,
		"iptables", "-t", "mangle", "-F",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{},
		"ip", "netns", "exec", new,
		"iptables", "-F",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{},
		"ip", "netns", "exec", new,
		"ip6tables", "-t", "nat", "-F",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{},
		"ip", "netns", "exec", new,
		"ip6tables", "-t", "mangle", "-F",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{},
		"ip", "netns", "exec", new,
		"ip6tables", "-F",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{},
		"ip", "netns", "exec", new,
		"ipset", "destroy",
	)
	if err != nil {
		return
	}

	ipsetData, err := utils.ExecOutputLogged(
		[]string{},
		"ip", "netns", "exec", cur,
		"ipset", "save",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecInputOutputLogged(
		[]string{},
		ipsetData,
		"ip", "netns", "exec", new,
		"ipset", "restore",
	)
	if err != nil {
		return
	}

	iptablesData, err := utils.ExecOutputLogged(
		[]string{},
		"ip", "netns", "exec", cur,
		"iptables-save",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecInputOutputLogged(
		[]string{},
		iptablesData,
		"ip", "netns", "exec", new,
		"iptables-restore",
	)
	if err != nil {
		return
	}

	ip6tablesData, err := utils.ExecOutputLogged(
		[]string{},
		"ip", "netns", "exec", cur,
		"ip6tables-save",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecInputOutputLogged(
		[]string{},
		ip6tablesData,
		"ip", "netns", "exec", new,
		"ip6tables-restore",
	)
	if err != nil {
		return
	}

	return
}

func Rename(cur, new string) (err error) {
	lock.Lock()
	defer lock.Unlock()

	err = CopyIptables(new, cur)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{
			"No such file",
		},
		"ip", "netns", "del", new,
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"/usr/bin/touch",
		path.Join("/", "run", "netns", path.Base(new)),
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"/usr/bin/mount",
		"--bind",
		path.Join("/", "run", "netns", path.Base(cur)),
		path.Join("/", "run", "netns", path.Base(new)),
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"/usr/bin/umount",
		path.Join("/", "run", "netns", path.Base(cur)),
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"/usr/bin/rm",
		path.Join("/", "run", "netns", path.Base(cur)),
	)
	if err != nil {
		return
	}

	return
}
