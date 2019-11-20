package cloudinit

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net"
	"net/textproto"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/authority"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
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
    mac_address: {{.Mac}}{{.Mtu}}
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

const netMtu = `
    mtu: %d`

const cloudConfigTmpl = `#cloud-config
ssh_deletekeys: false
disable_root: true
ssh_pwauth: no
growpart:
    mode: auto
    devices: ["/"]
    ignore_growroot_disabled: false
runcmd:
  - [ sysctl, -w, net.ipv4.conf.eth0.send_redirects=0 ]
users:
  - name: root
    lock-passwd: true
  - name: cloud
    groups: adm, video, wheel, systemd-journal
    selinux-user: staff_u
    sudo: ALL=(ALL) NOPASSWD:ALL
    lock-passwd: {{.LockPasswd}}
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
	Mtu      string
	Address  string
	Netmask  string
	Network  string
	Gateway  string
	Address6 string
	Gateway6 string
}

type cloudConfigData struct {
	LockPasswd string
	Keys       []string
}

func getUserData(db *database.Database, inst *instance.Instance,
	virt *vm.VirtualMachine, initial bool) (usrData string, err error) {

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

	if !initial {
		data.LockPasswd = "false"
	} else {
		data.LockPasswd = "true"
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

	if adapter.Vpc.IsZero() {
		err = &errortypes.NotFoundError{
			errors.Wrap(err, "cloudinit: Instance missing VPC"),
		}
		return
	}

	if adapter.Subnet.IsZero() {
		err = &errortypes.NotFoundError{
			errors.Wrap(err, "cloudinit: Instance missing VPC subnet"),
		}
		return
	}

	zne, err := zone.Get(db, node.Self.Zone)
	if err != nil {
		return
	}

	vxlan := false
	if zne.NetworkMode == zone.VxlanVlan {
		vxlan = true
	}

	vc, err := vpc.Get(db, adapter.Vpc)
	if err != nil {
		return
	}

	vcNet, err := vc.GetNetwork()
	if err != nil {
		return
	}

	addr, gatewayAddr, err := vc.GetIp(db, inst.Subnet, inst.Id)
	if err != nil {
		return
	}

	addr6 := vc.GetIp6(addr)
	gatewayAddr6 := vc.GetIp6(gatewayAddr)

	data := netConfigData{
		Mac:      adapter.MacAddress,
		Address:  addr.String(),
		Netmask:  net.IP(vcNet.Mask).String(),
		Network:  vcNet.IP.String(),
		Gateway:  gatewayAddr.String(),
		Address6: addr6.String(),
		Gateway6: gatewayAddr6.String(),
	}

	jumboFrames := node.Self.JumboFrames
	if jumboFrames || vxlan {
		mtuSize := 0
		if jumboFrames {
			mtuSize = settings.Hypervisor.JumboMtu
		} else {
			mtuSize = settings.Hypervisor.NormalMtu
		}

		if vxlan {
			mtuSize -= 54
		}

		data.Mtu = fmt.Sprintf(netMtu, mtuSize)
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
	virt *vm.VirtualMachine, initial bool) (err error) {

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

	usrData, err := getUserData(db, inst, virt, initial)
	if err != nil {
		return
	}

	metaData := fmt.Sprintf(metaDataTmpl,
		primitive.NewObjectID().Hex(),
		strings.Replace(inst.Name, " ", "_", -1),
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
