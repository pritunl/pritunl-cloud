package sysctl

import (
	"github.com/pritunl/pritunl-cloud/builder/prompt"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

const (
	securityLimits = `* soft memlock 2048000000
* hard memlock 2048000000
root soft memlock 2048000000
root hard memlock 2048000000
* hard nofile 500000
* soft nofile 500000
root hard nofile 500000
root soft nofile 500000
`
	dirtyRatio = `vm.dirty_ratio = 5
vm.dirty_background_ratio = 3
`
	raidSpeedLimit = `dev.raid.speed_limit_max = 100000
`
	schedulerMigration = `kernel.sched_migration_cost_ns = 5000000
`
	cachePressure = `vm.vfs_cache_pressure = 200
`
	selinuxConfDisabled = `
# This file controls the state of SELinux on the system.
# SELINUX= can take one of these three values:
#     enforcing - SELinux security policy is enforced.
#     permissive - SELinux prints warnings instead of enforcing.
#     disabled - No SELinux policy is loaded.
SELINUX=disabled
# SELINUXTYPE= can take one of these three values:
#     targeted - Targeted processes are protected,
#     minimum - Modification of targeted policy. Only selected processes are protected. 
#     mls - Multi Level Security protection.
SELINUXTYPE=targeted


`
	selinuxConfEnabled = `
# This file controls the state of SELinux on the system.
# SELINUX= can take one of these three values:
#     enforcing - SELinux security policy is enforced.
#     permissive - SELinux prints warnings instead of enforcing.
#     disabled - No SELinux policy is loaded.
SELINUX=enforcing
# SELINUXTYPE= can take one of these three values:
#     targeted - Targeted processes are protected,
#     minimum - Modification of targeted policy. Only selected processes are protected. 
#     mls - Multi Level Security protection.
SELINUXTYPE=targeted


`
)

func OpenFileLimit() (err error) {
	resp, err := prompt.ConfirmDefault(
		"Increase open file limit [Y/n]",
		true,
	)
	if err != nil {
		return
	}

	if !resp {
		return
	}

	pth := "/etc/security/limits.conf"
	exists, err := utils.Exists(pth)
	if err != nil {
		return
	}

	if exists {
		err = utils.Write(
			pth,
			securityLimits,
			0644,
		)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"path": pth,
		}).Info("sysctl: Increased open file limit")
	}

	return
}

func DirtyRatio() (err error) {
	resp, err := prompt.ConfirmDefault(
		"Reduce dirty memory limit [Y/n]",
		true,
	)
	if err != nil {
		return
	}

	pth := "/etc/sysctl.d/10-dirty.conf"
	if resp {
		err = utils.CreateWrite(
			pth,
			dirtyRatio,
			0644,
		)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"path": pth,
		}).Info("sysctl: Dirty memory limit enabled")
	} else {
		exists, e := utils.Exists(pth)
		if e != nil {
			err = e
			return
		}

		if exists {
			err = utils.Remove(pth)
			if err != nil {
				return
			}

			logrus.WithFields(logrus.Fields{
				"path": pth,
			}).Info("sysctl: Dirty memory limit disabled")
		}
	}

	return
}

func SchedulerMigration() (err error) {
	resp, err := prompt.ConfirmDefault(
		"Optimize scheduler migration [Y/n]",
		true,
	)
	if err != nil {
		return
	}

	pth := "/etc/sysctl.d/10-scheduler.conf"
	if resp {
		err = utils.CreateWrite(
			pth,
			schedulerMigration,
			0644,
		)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"path": pth,
		}).Info("sysctl: Optimize scheduler migration enabled")
	} else {
		exists, e := utils.Exists(pth)
		if e != nil {
			err = e
			return
		}

		if exists {
			err = utils.Remove(pth)
			if err != nil {
				return
			}

			logrus.WithFields(logrus.Fields{
				"path": pth,
			}).Info("sysctl: Optimize scheduler migration disabled")
		}
	}

	return
}

func CachePressure() (err error) {
	resp, err := prompt.ConfirmDefault(
		"Optimize cache pressure [Y/n]",
		true,
	)
	if err != nil {
		return
	}

	pth := "/etc/sysctl.d/10-cache.conf"
	if resp {
		err = utils.CreateWrite(
			pth,
			cachePressure,
			0644,
		)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"path": pth,
		}).Info("sysctl: Optimize cache pressure enabled")
	} else {
		exists, e := utils.Exists(pth)
		if e != nil {
			err = e
			return
		}

		if exists {
			err = utils.Remove(pth)
			if err != nil {
				return
			}

			logrus.WithFields(logrus.Fields{
				"path": pth,
			}).Info("sysctl: Optimize cache pressure disabled")
		}
	}

	return
}

func RaidSpeedLimit() (err error) {
	resp, err := prompt.ConfirmDefault(
		"Limit raid sync speed [Y/n]",
		true,
	)
	if err != nil {
		return
	}

	pth := "/etc/sysctl.d/10-raid.conf"
	if resp {
		err = utils.CreateWrite(
			pth,
			raidSpeedLimit,
			0644,
		)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"path": pth,
		}).Info("sysctl: Raid sync speed limit enabled")
	} else {
		exists, e := utils.Exists(pth)
		if e != nil {
			err = e
			return
		}

		if exists {
			err = utils.Remove(pth)
			if err != nil {
				return
			}

			logrus.WithFields(logrus.Fields{
				"path": pth,
			}).Info("sysctl: Raid sync speed limit disabled")
		}
	}

	return
}

func Selinux() (err error) {
	resp, err := prompt.ConfirmDefault(
		"Disable SELinux (required for Pritunl Cloud) [Y/n]",
		true,
	)
	if err != nil {
		return
	}

	pth := "/etc/selinux/config"
	if resp {
		err = utils.CreateWrite(
			pth,
			selinuxConfDisabled,
			0644,
		)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"path": pth,
		}).Info("sysctl: SELinux disabled")
	} else {
		err = utils.CreateWrite(
			pth,
			selinuxConfEnabled,
			0644,
		)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"path": pth,
		}).Info("sysctl: SELinux enabled")
	}

	pth = "/etc/sysconfig/selinux"
	if resp {
		err = utils.CreateWrite(
			pth,
			selinuxConfDisabled,
			0644,
		)
		if err != nil {
			return
		}

		utils.ExecCombinedOutput("", "setenforce", "0")

		logrus.WithFields(logrus.Fields{
			"path": pth,
		}).Info("sysctl: SELinux disabled")
	} else {
		err = utils.CreateWrite(
			pth,
			selinuxConfEnabled,
			0644,
		)
		if err != nil {
			return
		}

		utils.ExecCombinedOutput("", "setenforce", "1")

		logrus.WithFields(logrus.Fields{
			"path": pth,
		}).Info("sysctl: SELinux enabled")
	}

	return
}

func Optimize() (err error) {
	err = OpenFileLimit()
	if err != nil {
		return
	}

	err = DirtyRatio()
	if err != nil {
		return
	}

	err = CachePressure()
	if err != nil {
		return
	}

	err = RaidSpeedLimit()
	if err != nil {
		return
	}

	err = Selinux()
	if err != nil {
		return
	}

	return
}
