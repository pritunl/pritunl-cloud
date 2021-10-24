package namespace

import (
	"path"
	"sync"

	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	lock = sync.Mutex{}
)

func Rename(cur, new string) (err error) {
	lock.Lock()
	defer lock.Unlock()

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
