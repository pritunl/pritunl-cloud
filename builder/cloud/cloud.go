package cloud

import (
	"encoding/json"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/builder/constants"
	"github.com/pritunl/pritunl-cloud/builder/prompt"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

const (
	repoPritunlKeyUrl      = "https://raw.githubusercontent.com/pritunl/pgp/master/pritunl_repo_pub.asc"
	repoPritunlKeyPath     = "/tmp/pritunl.pub"
	repoPritunlFingerprint = "7568D9BB55FF9E5287D586017AE645C0CF8E292A"
	repoPritunlPath        = "/etc/yum.repos.d/pritunl.repo"
	repoPritunlData        = `[pritunl]
name=Pritunl Repository
baseurl=https://repo.pritunl.com/stable/yum/oraclelinux/9/
gpgcheck=1
enabled=1
`
	repoPritunlUnstableData = `[pritunl]
name=Pritunl Unstable Repository
baseurl=https://repo.pritunl.com/unstable/yum/oraclelinux/9/
gpgcheck=1
enabled=1
`
	repoKvmKeyUrl      = "https://raw.githubusercontent.com/pritunl/pgp/master/pritunl_repo_pub.asc"
	repoKvmKeyPath     = "/tmp/pritunl-kvm.pub"
	repoKvmFingerprint = "1BB6FBB8D641BD9C6C0398D74D55437EC0508F5F"
	repoKvmPath        = "/etc/yum.repos.d/pritunl-kvm.repo"
	repoKvmData        = `[pritunl-kvm]
name=Pritunl KVM Repository
baseurl=https://repo.pritunl.com/kvm-stable/oraclelinux/9/
gpgcheck=1
enabled=1
`
	repoKvmUnstableData = `[pritunl-kvm]
name=Pritunl KVM Repository
baseurl=https://repo.pritunl.com/kvm-unstable/oraclelinux/9/
gpgcheck=1
enabled=1
`
	aptRepoPritunlPath = "/etc/apt/sources.list.d/pritunl.list"
	aptRepoPritunlData = `deb http://repo.pritunl.com/stable/apt focal main
`
	aptRepoPritunlUnstableData = `deb http://repo.pritunl.com/unstable/apt focal main
`
	aptRepoKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

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
=lhbs
-----END PGP PUBLIC KEY BLOCK-----
`
)

type InstallConfigData struct {
	MongoUri string `json:"mongo_uri"`
}

func KvmRepo(unstable bool) (err error) {
	if constants.Target != constants.Rpm {
		return
	}

	if unstable {
		err = utils.CreateWrite(
			repoKvmPath,
			repoKvmUnstableData,
			0644,
		)
	} else {
		err = utils.CreateWrite(
			repoKvmPath,
			repoKvmData,
			0644,
		)
	}
	if err != nil {
		return
	}

	//err = utils.Exec("",
	//	"/usr/bin/gpg",
	//	"--keyserver", "hkp://keyserver.ubuntu.com",
	//	"--recv-keys", repoKvmFingerprint,
	//)
	//if err != nil {
	//	return
	//}
	//
	//err = utils.Exec("",
	//	"/usr/bin/gpg",
	//	"--armor",
	//	"--output", repoKvmKeyPath,
	//	"--export", repoKvmFingerprint,
	//)
	//if err != nil {
	//	return
	//}
	//
	//err = utils.Exec("",
	//	"/usr/bin/rpm",
	//	"--import", repoKvmKeyPath,
	//)
	//if err != nil {
	//	return
	//}
	//
	//_ = os.Remove(repoKvmKeyPath)

	err = utils.Exec("",
		"/usr/bin/rpm",
		"--import", repoKvmKeyUrl,
	)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"path": repoKvmPath,
	}).Info("cloud: Pritunl KVM repository added")

	return
}

func pritunlRepoRpm(unstable bool) (err error) {
	if unstable {
		err = utils.CreateWrite(
			repoPritunlPath,
			repoPritunlUnstableData,
			0644,
		)
		if err != nil {
			return
		}
	} else {
		err = utils.CreateWrite(
			repoPritunlPath,
			repoPritunlData,
			0644,
		)
		if err != nil {
			return
		}
	}

	//err = utils.Exec("",
	//	"/usr/bin/gpg",
	//	"--keyserver", "hkp://keyserver.ubuntu.com",
	//	"--recv-keys", repoPritunlFingerprint,
	//)
	//if err != nil {
	//	return
	//}
	//
	//err = utils.Exec("",
	//	"/usr/bin/gpg",
	//	"--armor",
	//	"--output", repoPritunlKeyPath,
	//	"--export", repoPritunlFingerprint,
	//)
	//if err != nil {
	//	return
	//}
	//
	//err = utils.Exec("",
	//	"/usr/bin/rpm",
	//	"--import", repoPritunlKeyPath,
	//)
	//if err != nil {
	//	return
	//}
	//
	//_ = os.Remove(repoPritunlKeyPath)

	err = utils.Exec("",
		"/usr/bin/rpm",
		"--import", repoPritunlKeyUrl,
	)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"path": repoPritunlPath,
	}).Info("cloud: Pritunl repository added")

	return
}

func pritunlRepoApt(unstable bool) (err error) {
	err = utils.ExecInput("",
		aptRepoKey,
		"/usr/bin/apt-key",
		"add", "-",
	)
	if err != nil {
		return
	}

	if unstable {
		err = utils.CreateWrite(
			aptRepoPritunlPath,
			aptRepoPritunlUnstableData,
			0644,
		)
		if err != nil {
			return
		}
	} else {
		err = utils.CreateWrite(
			aptRepoPritunlPath,
			aptRepoPritunlData,
			0644,
		)
		if err != nil {
			return
		}
	}

	logrus.WithFields(logrus.Fields{
		"path": aptRepoPritunlPath,
	}).Info("cloud: Pritunl repository added")

	return
}

func PritunlRepo(unstable bool) error {
	if constants.Target == constants.Apt {
		return pritunlRepoApt(unstable)
	} else {
		return pritunlRepoRpm(unstable)
	}
}

func installRpm() (err error) {
	utils.ExecCombinedOutput("", "/usr/bin/systemctl", "stop", "libvirtd")
	utils.ExecCombinedOutput("", "/usr/bin/systemctl", "disable", "libvirtd")

	err = utils.Exec("", "/usr/bin/yum", "-y", "install",
		"edk2-ovmf", "qemu-kvm", "qemu-kvm-core", "qemu-img", "genisoimage",
		"swtpm", "pritunl-cloud")
	if err != nil {
		return
	}

	_, err = utils.ExecOutputLogged(nil,
		"/usr/bin/systemctl", "enable", "pritunl-cloud")
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"package": "pritunl-cloud",
	}).Info("cloud: Pritunl Cloud install")

	return
}

func installApt() (err error) {
	err = utils.Exec("", "/usr/bin/apt", "update")
	if err != nil {
		return
	}

	err = utils.Exec("", "/usr/bin/apt", "-y", "install",
		"ovmf", "qemu", "qemu-efi", "qemu-system-x86", "qemu-user",
		"swtpm", "qemu-utils", "genisoimage", "pritunl-cloud")
	if err != nil {
		return
	}

	utils.ExecCombinedOutput("", "/usr/bin/systemctl", "stop", "libvirtd")
	utils.ExecCombinedOutput("", "/usr/bin/systemctl", "disable", "libvirtd")

	_, err = utils.ExecOutputLogged(nil,
		"/usr/bin/systemctl", "enable", "pritunl-cloud")
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"package": "pritunl-cloud",
	}).Info("cloud: Pritunl Cloud install")

	return
}

func Install() error {
	if constants.Target == constants.Apt {
		return installApt()
	} else {
		return installRpm()
	}
}

func InstallKvm() (err error) {
	err = utils.Exec("", "/usr/bin/yum", "-y", "remove",
		"qemu-kvm", "qemu-img", "qemu-system-x86", "cockpit", "cockpit-ws")
	if err != nil {
		return
	}

	utils.ExecCombinedOutput("", "/usr/bin/systemctl", "stop", "libvirtd")
	utils.ExecCombinedOutput("", "/usr/bin/systemctl", "disable", "libvirtd")

	err = utils.Exec("", "/usr/bin/yum", "-y", "install",
		"edk2-ovmf", "genisoimage", "swtpm", "pritunl-qemu-kvm",
		"pritunl-qemu-img", "pritunl-qemu-system-x86", "pritunl-cloud")
	if err != nil {
		return
	}

	_, err = utils.ExecOutputLogged(nil,
		"/usr/bin/systemctl", "enable", "pritunl-cloud")
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"package": "pritunl-cloud",
	}).Info("cloud: Pritunl Cloud install")

	return
}

func InstallMongo(mongoUri string) (err error) {
	confData, err := json.Marshal(&InstallConfigData{
		MongoUri: mongoUri,
	})
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "builder: Failed to marshal config"),
		}
		return
	}

	err = utils.CreateWrite(
		"/etc/pritunl-cloud.json",
		string(confData),
		0600,
	)
	if err != nil {
		return
	}

	return
}

func Cloud(unstable bool) (err error) {
	kvmResp := false
	if constants.Target == constants.Rpm {
		kvmResp, err = prompt.ConfirmDefault(
			"Enable Pritunl Cloud KVM Repo [Y/n]",
			true,
		)
		if err != nil {
			return
		}
	}

	if kvmResp {
		err = KvmRepo(unstable)
		if err != nil {
			return
		}
	}

	cloudResp, err := prompt.ConfirmDefault(
		"Install Pritunl Cloud [Y/n]",
		true,
	)
	if err != nil {
		return
	}

	if !cloudResp {
		return
	}

	err = PritunlRepo(unstable)
	if err != nil {
		return
	}

	if kvmResp {
		err = InstallKvm()
		if err != nil {
			return
		}
	} else {
		err = Install()
		if err != nil {
			return
		}
	}

	mongoUri, err := prompt.InputDefault(
		"Enter MongoDB URI [mongodb://localhost:27017/pritunl-cloud]",
		"",
	)
	if err != nil {
		return
	}

	if mongoUri != "" {
		err = InstallMongo(mongoUri)
		if err != nil {
			return
		}
	}

	return
}
