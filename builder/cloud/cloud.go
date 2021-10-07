package cloud

import (
	"os"

	"github.com/pritunl/pritunl-cloud/builder/prompt"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

const (
	repoPritunlKeyPath     = "/tmp/pritunl.pub"
	repoPritunlFingerprint = "7568D9BB55FF9E5287D586017AE645C0CF8E292A"
	repoPritunlPath        = "/etc/yum.repos.d/pritunl.repo"
	repoPritunlData        = `[pritunl]
name=Pritunl Repository
baseurl=https://repo.pritunl.com/stable/yum/oraclelinux/8/
gpgcheck=1
enabled=1
`
	repoKvmKeyPath     = "/tmp/pritunl-kvm.pub"
	repoKvmFingerprint = "1BB6FBB8D641BD9C6C0398D74D55437EC0508F5F"
	repoKvmPath        = "/etc/yum.repos.d/pritunl-kvm.repo"
	repoKvmData        = `[pritunl-kvm]
name=Pritunl KVM Repository
baseurl=https://repo.pritunl.com/kvm/
gpgcheck=1
enabled=1
`
)

func KvmRepo() (err error) {
	err = utils.CreateWrite(
		repoKvmPath,
		repoKvmData,
		0644,
	)
	if err != nil {
		return
	}

	err = utils.Exec("",
		"/usr/bin/gpg",
		"--keyserver", "hkp://keyserver.ubuntu.com",
		"--recv-keys", repoKvmFingerprint,
	)
	if err != nil {
		return
	}

	err = utils.Exec("",
		"/usr/bin/gpg",
		"--armor",
		"--output", repoKvmKeyPath,
		"--export", repoKvmFingerprint,
	)
	if err != nil {
		return
	}

	err = utils.Exec("",
		"/usr/bin/rpm",
		"--import", repoKvmKeyPath,
	)
	if err != nil {
		return
	}

	_ = os.Remove(repoKvmKeyPath)

	logrus.WithFields(logrus.Fields{
		"path": repoKvmPath,
	}).Info("cloud: Pritunl KVM repository added")

	return
}

func PritunlRepo() (err error) {
	err = utils.CreateWrite(
		repoPritunlPath,
		repoPritunlData,
		0644,
	)
	if err != nil {
		return
	}

	err = utils.Exec("",
		"/usr/bin/gpg",
		"--keyserver", "hkp://keyserver.ubuntu.com",
		"--recv-keys", repoPritunlFingerprint,
	)
	if err != nil {
		return
	}

	err = utils.Exec("",
		"/usr/bin/gpg",
		"--armor",
		"--output", repoPritunlKeyPath,
		"--export", repoPritunlFingerprint,
	)
	if err != nil {
		return
	}

	err = utils.Exec("",
		"/usr/bin/rpm",
		"--import", repoPritunlKeyPath,
	)
	if err != nil {
		return
	}

	_ = os.Remove(repoPritunlKeyPath)

	logrus.WithFields(logrus.Fields{
		"path": repoPritunlPath,
	}).Info("cloud: Pritunl repository added")

	return
}

func Install() (err error) {
	err = utils.Exec("", "/usr/bin/yum", "-y", "remove",
		"qemu-kvm", "qemu-img", "qemu-system-x86", "cockpit", "cockpit-ws")
	if err != nil {
		return
	}

	utils.ExecOutputLogged(nil,
		"/usr/bin/systemctl", "stop", "libvirtd")
	utils.ExecOutputLogged(nil,
		"/usr/bin/systemctl", "disable", "libvirtd")

	err = utils.Exec("", "/usr/bin/yum", "-y", "install",
		"edk2-ovmf", "pritunl-qemu-kvm", "pritunl-qemu-img",
		"pritunl-qemu-system-x86", "pritunl-cloud", "genisoimage", "libusal")
	if err != nil {
		return
	}

	_, err = utils.ExecOutputLogged(nil,
		"/usr/bin/systemctl", "enable", "pritunl-cloud")
	if err != nil {
		return
	}

	_, err = utils.ExecOutputLogged(nil,
		"/usr/bin/systemctl", "start", "pritunl-cloud")
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"package": "pritunl-cloud",
	}).Info("mongo: Pritunl Cloud install")

	return
}

func Cloud() (err error) {
	resp, err := prompt.ConfirmDefault(
		"Install Pritunl Cloud [Y/n]",
		true,
	)
	if err != nil {
		return
	}

	if !resp {
		return
	}

	err = KvmRepo()
	if err != nil {
		return
	}

	err = PritunlRepo()
	if err != nil {
		return
	}

	err = Install()
	if err != nil {
		return
	}

	return
}
