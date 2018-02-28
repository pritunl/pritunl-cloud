package cloudinit

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/authority"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
	"mime/multipart"
	"net/textproto"
	"os"
	"path"
	"strings"
)

const metaDataTmpl = `instance-id: %s
local-hostname: %s`

const userDataTmpl = `Content-Type: multipart/mixed; boundary="%s"
MIME-Version: 1.0

%s`

const cloudConfigTmpl = `#cloud-config
ssh_deletekeys: true
disable_root: true
ssh_pwauth: no
users:
  - name: cloud
    groups: adm, video, wheel, systemd-journal
    selinux-user: staff_u
    sudo: ALL=(ALL) NOPASSWD:ALL
    lock-passwd: true
    ssh-authorized-keys:
%s`

const cloudScriptTmpl = `#!/bin/bash
%s`

const teeTmpl = `sudo tee %s << EOF
%s
EOF
`

func getUserData(db *database.Database, instId bson.ObjectId) (
	usrData string, err error) {

	inst, err := instance.Get(db, instId)
	if err != nil {
		return
	}

	authrs, err := authority.GetOrgRoles(db, inst.Organization,
		inst.NetworkRoles)
	if err != nil {
		return
	}

	if len(authrs) == 0 {
		return
	}

	authorizedKeys := ""
	trusted := ""
	principals := ""

	for _, authr := range authrs {
		switch authr.Type {
		case authority.SshKey:
			for _, key := range strings.Split(authr.Key, "\n") {
				authorizedKeys += fmt.Sprintf("      - %s\n", key)
			}
			break
		case authority.SshCertificate:
			trusted += authr.Certificate + "\n"
			principals += strings.Join(authr.Roles, "\n") + "\n"
			break
		}
	}

	items := []string{}

	items = append(items, fmt.Sprintf(cloudConfigTmpl, authorizedKeys))

	cloudScript := ""
	if trusted != "" {
		cloudScript += fmt.Sprintf(teeTmpl, "/etc/ssh/trusted", trusted)
	}
	if principals != "" {
		cloudScript += fmt.Sprintf(teeTmpl, "/etc/ssh/principals", principals)
	}

	if cloudScript != "" {
		items = append(items, fmt.Sprintf(cloudScriptTmpl, cloudScript))
	}

	buffer := &bytes.Buffer{}
	message := multipart.NewWriter(buffer)
	for _, item := range items {
		header := textproto.MIMEHeader{}

		header.Set("Content-Transfer-Encoding", "base64")
		header.Set("MIME-Version", "1.0")

		if strings.HasPrefix(item, "#!") {
			header.Set("Content-Type",
				"text/x-shellscript; charset=\"utf-8\"")
		} else {
			header.Set("Content-Type",
				"text/cloud-config; charset=\"utf-8\"")
		}

		part, e := message.CreatePart(header)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "cloudinit: Failed to create part"),
			}
			return
		}

		_, err = part.Write(
			[]byte(base64.StdEncoding.EncodeToString([]byte(item)) + "\n"))
		if err != nil {
			err = &errortypes.WriteError{
				errors.Wrap(err, "cloudinit: Failed to write part"),
			}
			return
		}
	}

	err = message.Close()
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "cloudinit: Failed to close message"),
		}
		return
	}

	usrData = fmt.Sprintf(
		userDataTmpl,
		message.Boundary(),
		buffer.String(),
	)

	return
}

func Write(db *database.Database, instId bson.ObjectId) (err error) {
	tempDir := paths.GetTempDir()
	metaPath := path.Join(tempDir, "meta-data")
	userPath := path.Join(tempDir, "user-data")
	initPath := paths.GetInitPath(instId)

	defer os.RemoveAll(tempDir)

	err = utils.ExistsMkdir(paths.GetInitsPath(), 0755)
	if err != nil {
		return
	}

	err = utils.ExistsMkdir(tempDir, 0700)
	if err != nil {
		return
	}

	usrData, err := getUserData(db, instId)
	if err != nil {
		return
	}

	metaData := fmt.Sprintf(metaDataTmpl, instId.Hex(), instId.Hex())

	err = utils.CreateWrite(metaPath, metaData, 0644)
	if err != nil {
		return
	}

	err = utils.CreateWrite(userPath, usrData, 0644)
	if err != nil {
		return
	}

	err = utils.Exec(tempDir,
		"genisoimage",
		"-output", initPath,
		"-volid", "cidata",
		"-joliet",
		"-rock",
		"user-data",
		"meta-data",
	)

	return
}
