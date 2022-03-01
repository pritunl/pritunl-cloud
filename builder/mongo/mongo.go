package mongo

import (
	"github.com/pritunl/pritunl-cloud/builder/constants"
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
	aptRepoPath = "/etc/apt/sources.list.d/mongodb-org-5.0.list"
	aptRepoData = `deb https://repo.mongodb.org/apt/ubuntu focal/mongodb-org/5.0 multiverse
`
	aptRepoKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mQINBGAsKNUBEAClMqPCvvqm6gFmbiorEN9qp00GI8oaECkwbxtGGbqX9sqMSrKe
AB3sGI7kqG2Fl0K+xmmiq1QDjhNgFDA1jjXq+Bd66RNPtvu747IRxVs+9fX7bk67
8Bruha7U3M5l4193x5oYLlbcZL9aC7RSJE2mggTyS6LarmF6vKQN9LMXDicnageV
KCPpF2i3jkZaGnLPzAisW/pOjPQpWCbatTVqKOKvtOyP3Fz1spYd4obu6ELu1PXa
gmhSfvWJYt1irpchOl29LWZfcmXuJszmb00bqm4gLcK12VrnK191iXv46A8h2hSO
f3eQqrkc+pF/kw4RyG54EV7QtHXyTe9TVCbJUfgtliWIQt/bCoJYfPLHJaWIMs83
bzA6ZvOjCKIfMS0CY5ZJyVaBfiI3wURSjgZIYFZAXVwbreQIfOKKuik7UVVn3xUO
nWpmQ2zyI0W7cJMquxwLNjkI+RckPhIqxWFo5iNSV4v6pzrlHD1WmIfFGBKEn7m+
edwVyHG53fNIFZjxyShO6Pf1vgb9Js/XmXB4lxYnNyx1tB+hQhXTjLlY6N5gPpw5
Z/PWQc7vfYekUZGQMXhTyRxU0QTwmdEeKcb+fb9r23OH59bbAfzE10xTMzhqCd2L
lgSozMBvMmkHb1xs1x6FFuv/U/X7LjHTrHIf4M//DNwdP4l4I1jhPlTAxwARAQAB
tDdNb25nb0RCIDUuMCBSZWxlYXNlIFNpZ25pbmcgS2V5IDxwYWNrYWdpbmdAbW9u
Z29kYi5jb20+iQI+BBMBAgAoBQJgLCjVAhsDBQkJZgGABgsJCAcDAgYVCAIJCgsE
FgIDAQIeAQIXgAAKCRCwCgvR4sY8EawdD/0ewkyx3yE99K9n3y7gdvh5+2U8BsqU
7SWEfup7kPpf+4pF5xWqMaciEV/wRAGt7TiKlfVyAv3Q9iNsaLFN+s3kMaIcKhwD
8+q/iGfziIuOSTeo20dAxn9vF6YqrKGc7TbHdXf9AtYuJCfIU5j02uVZiupx+P9+
rG39dEnjOXm3uY0Fv3pRGCpuGubDlWB1DYh0R5O481kDVGoMqBxmc3iTALu14L/u
g+AKxFYfT4DmgdzPVMDhppgywfyd/IOWxoOCl4laEhVjUt5CygBa7w07qdKwWx2w
gTd9U0KGHxnnSmvQYxrRrS5RX3ILPJShivTSZG+rMqnUe6RgCwBrKHCRU1L728Yv
1B3ZFJLxB1TlVT2Hjr+oigp0RY9W1FCIdO2uhb9GImpaJ1Y0ZZqUkt/d9D8U2wcw
SW6/6WYeO7wAi/zlJ25hrBwhxS2+88gM6wJ1yL9yrM9v8JUb7Kq0rCGsEO5kqscV
AmX90wsF2cZ6gHR53eGIDbAJK0MO5RHR73aQ4bpTivPnoTx4HTj5fyhW9z8yCSOe
BlQABoFFqFvOS7KBxoyIS3pxlDetWOSc6yQrvA1CwxnkB81OHNmJfWAbNbEtZkLm
xs2c8CIh2R81yi6HUzAaxyDH7mrThbwX3hUe/wsaD1koV91G6bDD4Xx3zpa9DG/O
HyB98+e983gslg==
=IQQF
-----END PGP PUBLIC KEY BLOCK-----
EOF
`
)

func repoRpm() (err error) {
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

func repoApt() (err error) {
	err = utils.ExecInput("",
		aptRepoKey,
		"/usr/bin/apt-key",
		"add", "-",
	)
	if err != nil {
		return
	}

	err = utils.CreateWrite(
		aptRepoPath,
		aptRepoData,
		0644,
	)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"path": aptRepoPath,
	}).Info("mongo: MongoDB repository added")

	return
}

func Repo() error {
	if constants.Target == constants.Apt {
		return repoApt()
	} else {
		return repoRpm()
	}
}

func installRpm() (err error) {
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

func installApt() (err error) {
	err = utils.Exec("", "/usr/bin/apt", "update")
	if err != nil {
		return
	}

	err = utils.Exec("", "/usr/bin/apt", "-y", "install", "mongodb-org")
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

func Install() error {
	if constants.Target == constants.Apt {
		return installApt()
	} else {
		return installRpm()
	}
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
