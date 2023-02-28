package mongo

import (
	"github.com/pritunl/pritunl-cloud/builder/constants"
	"github.com/pritunl/pritunl-cloud/builder/prompt"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

const (
	repoPath = "/etc/yum.repos.d/mongodb-org.repo"
	repoData = `[mongodb-org-6.0]
name=MongoDB Repository
baseurl=https://repo.mongodb.org/yum/redhat/9/mongodb-org/6.0/x86_64/
gpgcheck=1
enabled=1
gpgkey=https://www.mongodb.org/static/pgp/server-6.0.asc
`
	aptRepoPath = "/etc/apt/sources.list.d/mongodb-org-6.0.list"
	aptRepoData = `deb https://repo.mongodb.org/apt/ubuntu jammy/mongodb-org/6.0 multiverse
`
	aptRepoKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mQINBGIWTroBEADgSBs1z1MC5Hog5yd2wYHskzPE0SOl9LGB35Xhw1894hrKsswp
AS7JnViltXE71iJMoAqepJBvfmZLOyQO0rXcLlHXExK/IctnosRqGQeyLxNZKS0h
e1xQYQrPCWRaHqseYLuJ5wME49aFQ2YS7caFowBvKjsT5AoT7B0uXDp6nHZDUQG2
MBZJqUKziVYYt7PARv81llDNKqPvLDSc2McL/2aa4mNR/pM5r8iQjACbSnj37ERm
zca2gJ0GzCeZSqfmjoF7I6Ez1Nc/2ge1+fZA24pDFg+7W25du3JIqbnpJQAK5TAz
7tVzvEKU8WT9aQW3G1e5ox3YtlRPTSrTxN9dzLh123NGCd0J9a4moFkZIr8HmySd
jkdz4V1pKv9aTOhLjQpF/bhRaUuNuGK7TV7ZzY+PCVE51fmJx2EX4Ck5c6sW03rJ
59KbrxeTq02AcIBTFUY0Mfh7nxvYvwvLI0OKBOqFGXi4hFXpV4uo0rDLe+tGLFDD
+HsajFUUyAlMETE80PXOuTs44TZiW+SGCTyP2Sm8TBIiacSqsGNsryjgEDaIG6c1
FB++njqTfGlyZujamYbF3s3wBK8nDBVRympJcsHjLqUhvbh1Bq4hyF2pxio93SgA
mPEm6kl0KBCqpJNZpAFSVHK8penQtQUa0jFQetYPDUFfgTsg7qdZDQNcUwARAQAB
tDdNb25nb0RCIDYuMCBSZWxlYXNlIFNpZ25pbmcgS2V5IDxwYWNrYWdpbmdAbW9u
Z29kYi5jb20+iQI+BBMBAgAoBQJiFk66AhsDBQkJZgGABgsJCAcDAgYVCAIJCgsE
FgIDAQIeAQIXgAAKCRBqJrGuZMPDiADhEACex1qu1HbVIeBwZO4GYYEc8OpswguI
LvTL1ufWMVbpSFkm0XDzx7JU0SewCEBzr7BTri2zjNaPm7RQHYFl1ztTnNvxrvzu
AUoj/BClAgQXujSuUcEu+uA9pBHObiLHAkYFy61EnKgXu2iTOMn7HqRvjvHZyOnr
5llGG2zUq8YbEVs4GTHVV9CjCWBkf78stdqEAPCH69DtR1Bv2jQfUslVSDKUnluX
feTRDgWXnIKo4ld6EoqtYurIbcJIGvXHbFx90PoZiPJXn+eTY+6HS3I/TXDGAOkF
xkgmVsPWcZvbU0dLXjAiTIADODyiEiZlonrxYXJztIs/KXLl5CnvAEeXKXACbgaN
nuIMKtprtrLvFDpXwfyI90He0Vv8iE1wXSLcuztT5R1h6NmisMz9oRYQL3hqsSEn
TjV+Ko34Kyo459Bs9PhJO0DcZGg+B8iU9TdJgfp1KEs2HJFAueVtYAUJ3y5+UJFn
AkQoD5CC0Y+93z0+nHQPvjyxQ/7swFWNtrumrthcpYbGMIKEWqaQoEz2My5gVXHh
v5pHEXxXiARNe44GsS8r+1DYQypDUAh5Tw9mQRagWuC5Dsaaqob5vCdcFEAgiK5W
a/coP3B6WzUoQE8NKa8qnKDvX5RU0dxG5oUre+PuOwiHpom9G+375YYkwIL9a6pE
RRM5efxf1F532A==
=Cc71
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
