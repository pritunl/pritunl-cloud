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
	"math/rand"
	"mime/multipart"
	"net"
	"net/textproto"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
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
{{range .Keys}}      - {{.}}
{{end}}network:
  version: 2
  ethernets:
    ens3:
      match:
        macaddress: {{.HostMac}}
      dhcp4: true{{range .Networks}}
    ens{{.Num}}:
      match:
        macaddress: {{.Mac}}
      addresses:
        - {{.Address}}{{end}}
`

const cloudScriptTmpl = `#!/bin/bash
%s
sudo systemctl restart network`

const teeTmpl = `sudo tee %s << EOF
%s
EOF
`

const networkTmpl = `sudo tee /etc/sysconfig/network-scripts/ifcfg-ens%d << EOF
DEVICE=ens%d
HWADDR=%s
ONBOOT=yes
TYPE=Ethernet
USERCTL=no
IPADDR="%s"
NETMASK="%s"
EOF
`

var (
	cloudConfig = template.Must(template.New("cloud").Parse(cloudConfigTmpl))
)

type cloudNetworkData struct {
	Num     int
	Mac     string
	Address string
}

type cloudConfigData struct {
	Keys     []string
	HostMac  string
	Networks []cloudNetworkData
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
		Keys:     []string{},
		HostMac:  virt.NetworkAdapters[0].MacAddress,
		Networks: []cloudNetworkData{},
	}

	for i, adapter := range virt.NetworkAdapters {
		if i == 0 || adapter.Type != vm.Vxlan {
			continue
		}

		vc, e := vpc.Get(db, adapter.VpcId)
		if e != nil {
			err = e
			return
		}

		vcNet, e := vc.GetNetwork()
		if e != nil {
			err = e
			return
		}

		ip := vcNet.IP
		n := rand.Intn(250) + 2
		for x := 0; x < n; x++ {
			utils.IncIpAddress(ip)
		}

		cidr, _ := vcNet.Mask.Size()

		data.Networks = append(data.Networks, cloudNetworkData{
			Num:     i + 3,
			Mac:     adapter.MacAddress,
			Address: ip.String() + "/" + strconv.Itoa(cidr),
		})

		cloudScript += fmt.Sprintf(networkTmpl, i+3, i+3,
			adapter.MacAddress, ip.String(), net.IP(vcNet.Mask).String())
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

func Write(db *database.Database, inst *instance.Instance,
	virt *vm.VirtualMachine) (err error) {

	tempDir := paths.GetTempDir()
	metaPath := path.Join(tempDir, "meta-data")
	userPath := path.Join(tempDir, "user-data")
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

	metaData := fmt.Sprintf(metaDataTmpl, inst.Id.Hex(), inst.Id.Hex())

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
