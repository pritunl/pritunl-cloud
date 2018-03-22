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
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"gopkg.in/mgo.v2/bson"
	"mime/multipart"
	"net"
	"net/textproto"
	"os"
	"path"
	"strings"
	"text/template"
)

const metaDataTmpl = `instance-id: %s
local-hostname: %s`

const userDataTmpl = `Content-Type: multipart/mixed; boundary="%s"
MIME-Version: 1.0

%s`

const netConfigTmpl = `version: 1
config:
  - type: physical
    name: eth0
    mac_address: {{.Mac}}
    subnets:
      - type: static
        address: {{.Address}}
        netmask: {{.Netmask}}
        network: {{.Network}}
        gateway: {{.Gateway}}
        dns_nameservers:
          - 8.8.8.8
          - 8.8.4.4
      - type: static
        address: {{.Address6}}
        gateway: {{.Gateway6}}
`

const cloudConfigTmpl = `#cloud-config
ssh_deletekeys: false
disable_root: true
ssh_pwauth: no
users:
  - name: cloud
    groups: adm, video, wheel, systemd-journal
    selinux-user: staff_u
    sudo: ALL=(ALL) NOPASSWD:ALL
    lock-passwd: true
    ssh-authorized-keys:
{{range .Keys}}      - {{.}}
{{end}}`

const cloudScriptTmpl = `#!/bin/bash
%s`

const teeTmpl = `sudo tee %s << EOF
%s
EOF
`

var (
	cloudConfig = template.Must(template.New("cloud").Parse(cloudConfigTmpl))
	netConfig   = template.Must(template.New("net").Parse(netConfigTmpl))
)

type netConfigData struct {
	Mac      string
	Address  string
	Netmask  string
	Network  string
	Gateway  string
	Address6 string
	Gateway6 string
}

type cloudConfigData struct {
	Keys []string
}

func getUserData(db *database.Database, inst *instance.Instance,
	virt *vm.VirtualMachine) (usrData string, err error) {

	authrs, err := authority.GetOrgRoles(db, inst.Organization,
		inst.NetworkRoles)
	if err != nil {
		return
	}

	if len(authrs) == 0 {
		return
	}

	trusted := ""
	principals := ""
	cloudScript := ""

	data := cloudConfigData{
		Keys: []string{},
	}

	for _, authr := range authrs {
		switch authr.Type {
		case authority.SshKey:
			for _, key := range strings.Split(authr.Key, "\n") {
				data.Keys = append(data.Keys, key)
			}
			break
		case authority.SshCertificate:
			trusted += authr.Certificate + "\n"
			principals += strings.Join(authr.Roles, "\n") + "\n"
			break
		}
	}

	items := []string{}

	output := &bytes.Buffer{}
	err = cloudConfig.Execute(output, data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "cloudinit: Failed to exec cloud template"),
		}
		return
	}
	items = append(items, output.String())

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

func getNetData(db *database.Database, inst *instance.Instance,
	virt *vm.VirtualMachine) (netData string, err error) {

	if len(virt.NetworkAdapters) == 0 {
		err = &errortypes.NotFoundError{
			errors.Wrap(err, "cloudinit: Instance missing network adapters"),
		}
		return
	}

	adapter := virt.NetworkAdapters[0]

	if adapter.VpcId == "" {
		err = &errortypes.NotFoundError{
			errors.Wrap(err, "cloudinit: Instance missing VPC"),
		}
		return
	}

	vc, err := vpc.Get(db, adapter.VpcId)
	if err != nil {
		return
	}

	vcNet, err := vc.GetNetwork()
	if err != nil {
		return
	}

	addr, err := vc.GetIp(db, vpc.Instance, inst.Id)
	if err != nil {
		return
	}

	gatewayAddr, err := vc.GetIp(db, vpc.Gateway, inst.Id)
	if err != nil {
		return
	}

	addr6 := vc.GetIp6(addr)
	if err != nil {
		return
	}

	gatewayAddr6 := vc.GetIp6(gatewayAddr)
	if err != nil {
		return
	}

	data := netConfigData{
		Mac:      adapter.MacAddress,
		Address:  addr.String(),
		Netmask:  net.IP(vcNet.Mask).String(),
		Network:  vcNet.IP.String(),
		Gateway:  gatewayAddr.String(),
		Address6: addr6.String(),
		Gateway6: gatewayAddr6.String(),
	}

	output := &bytes.Buffer{}
	err = netConfig.Execute(output, data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "cloudinit: Failed to exec cloud template"),
		}
		return
	}

	netData = output.String()

	return
}

func Write(db *database.Database, inst *instance.Instance,
	virt *vm.VirtualMachine) (err error) {

	tempDir := paths.GetTempDir()
	metaPath := path.Join(tempDir, "meta-data")
	userPath := path.Join(tempDir, "user-data")
	netPath := path.Join(tempDir, "network-config")
	initPath := paths.GetInitPath(inst.Id)

	defer os.RemoveAll(tempDir)

	err = utils.ExistsMkdir(paths.GetInitsPath(), 0755)
	if err != nil {
		return
	}

	err = utils.ExistsMkdir(tempDir, 0700)
	if err != nil {
		return
	}

	usrData, err := getUserData(db, inst, virt)
	if err != nil {
		return
	}

	metaData := fmt.Sprintf(metaDataTmpl,
		bson.NewObjectId().Hex(),
		inst.Id.Hex(),
	)

	err = utils.CreateWrite(metaPath, metaData, 0644)
	if err != nil {
		return
	}

	netData, err := getNetData(db, inst, virt)
	if err != nil {
		return
	}

	err = utils.CreateWrite(netPath, netData, 0644)
	if err != nil {
		return
	}

	err = utils.CreateWrite(userPath, usrData, 0644)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLoggedDir(
		nil, tempDir,
		"genisoimage",
		"-output", initPath,
		"-volid", "cidata",
		"-joliet",
		"-rock",
		"user-data",
		"meta-data",
		"network-config",
	)

	return
}
