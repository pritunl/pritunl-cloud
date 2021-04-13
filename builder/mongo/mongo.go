package mongo

import (
	"github.com/pritunl/pritunl-cloud/builder/prompt"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

const (
	repoPath = "/etc/yum.repos.d/mongodb-org.repo"
	repoData = `[mongodb-org-4.4]
name=MongoDB Repository
baseurl=https://repo.mongodb.org/yum/redhat/8/mongodb-org/4.4/x86_64/
gpgcheck=1
enabled=1
gpgkey=https://www.mongodb.org/static/pgp/server-4.4.asc
`
)

func Repo() (err error) {
	err = utils.CreateWrite(
		repoPath,
		repoData,
		0644,
	)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"path": repoPath,
	}).Info("mongo: MongoDB repository added")

	return
}

func Install() (err error) {
	err = utils.Exec("", "/usr/bin/yum", "-y", "install", "mongodb-org")
	if err != nil {
		return
	}

	_, err = utils.ExecOutputLogged(nil,
		"/usr/bin/systemctl", "enable", "mongod")
	if err != nil {
		return
	}

	_, err = utils.ExecOutputLogged(nil,
		"/usr/bin/systemctl", "start", "mongod")
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"package": "mongodb-org",
	}).Info("mongo: MongoDB install")

	return
}

func Mongo() (err error) {
	resp, err := prompt.ConfirmDefault(
		"Install MongoDB [Y/n]",
		true,
	)
	if err != nil {
		return
	}

	if !resp {
		return
	}

	err = Repo()
	if err != nil {
		return
	}

	err = Install()
	if err != nil {
		return
	}

	return
}
