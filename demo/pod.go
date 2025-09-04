package demo

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/unit"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Pods = []*aggregate.PodAggregate{
	{
		Pod: pod.Pod{
			Id:               utils.ObjectIdHex("688bf358d978631566998ffc"),
			Name:             "web-app",
			Comment:          "",
			Organization:     utils.ObjectIdHex("688ab80d1793930f821f4f2c"),
			DeleteProtection: false,
			Drafts:           []*pod.UnitDraft{},
		},
		Units: Units,
	},
}

var Units = []*unit.Unit{
	{
		Id:           utils.ObjectIdHex("688c716d9da165ffad4b3682"),
		Pod:          utils.ObjectIdHex("688bf358d978631566998ffc"),
		Organization: utils.ObjectIdHex("688ab80d1793930f821f4f2c"),
		Name:         "web-app",
		Kind:         "instance",
		Count:        0,
		Deployments: []primitive.ObjectID{
			utils.ObjectIdHex("688db9219da165ffad4e439c"),
			utils.ObjectIdHex("688dbc759da165ffad4e4ab0"),
		},
		Spec: "```yaml" + `
---
name: web-app
kind: instance
zone: +/zone/us-west-1a
shape: +/shape/m2-small
processors: 4
memory: 4096
vpc: +/vpc/vpc
subnet: +/subnet/primary
image: +/image/almalinux9
roles:
    - instance
nodePorts:
    - protocol: tcp
      externalPort: 32120
      internalPort: 80

---
name: web-app-firewall
kind: firewall
ingress:
    - protocol: tcp
      port: 22
      source:
        - 10.20.0.0/16
` + "```" + `

## Initialization

* Update system
* Install nginx

` + "```shell" + `
dnf -y update
dnf install -y nginx
sed -i "s/Test Page/cloud-$(pci get +/instance/self/id)/" /usr/share/nginx/html/index.html
` + "```" + `

## Configuration

` + "```python {phase=reload}" + `
import string
import subprocess

def pci_get(query):
    return subprocess.run(
        ["pci", "get", query],
        stdout=subprocess.PIPE,
        text=True,
        check=True,
    ).stdout.strip()

with open("/etc/web.conf", "w") as file:
    file.write(pci_get("+/unit/database/private_ips"))
` + "```" + `

## Startup

` + "```shell {phase=reboot}" + `
systemctl start nginx
` + "```",
		SpecIndex:  2,
		LastSpec:   primitive.ObjectID{},
		DeploySpec: primitive.ObjectID{},
		Hash:       "80309b44139a78378c9025d40535c73f52f9d71c",
	},
	{
		Id:           utils.ObjectIdHex("68b67d1aee12c08a1f39f88b"),
		Pod:          utils.ObjectIdHex("688bf358d978631566998ffc"),
		Organization: utils.ObjectIdHex("688ab80d1793930f821f4f2c"),
		Name:         "database",
		Kind:         "instance",
		Count:        0,
		Deployments:  []primitive.ObjectID{},
		Spec: "```yaml" + `
---
name: database
kind: instance
zone: +/zone/us-west-1a
shape: +/shape/m2-small
processors: 4
memory: 4096
vpc: +/vpc/vpc
subnet: +/subnet/primary
image: +/image/almalinux9
roles:
    - instance

---
name: web-app-firewall
kind: firewall
ingress:
    - protocol: tcp
      port: 27017
      source:
        - +/unit/web-app
` + "```" + `

## Initialization

* Update system
* Install mongodb

` + "```shell" + `
dnf -y update

tee /etc/yum.repos.d/mongodb-org.repo << EOF
[mongodb-org-8.0]
name=MongoDB Repository
baseurl=https://repo.mongodb.org/yum/redhat/9/mongodb-org/8.0/x86_64/
gpgcheck=1
enabled=1
gpgkey=https://pgp.mongodb.com/server-8.0.asc
EOF

dnf -y install mongodb-org
` + "```" + `

## Startup

` + "```shell {phase=reboot}" + `
sudo systemctl start mongod
` + "```",
		SpecIndex:  2,
		LastSpec:   primitive.ObjectID{},
		DeploySpec: primitive.ObjectID{},
		Hash:       "d0c176ab5dafce10956e0d3d5e1320b2e496acff",
	},
}
