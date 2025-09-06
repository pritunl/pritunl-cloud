package demo

import (
	"time"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/unit"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Pods = []*aggregate.PodAggregate{
	{
		Pod: pod.Pod{
			Id:               utils.ObjectIdHex("688bf358d978631566998ffc"),
			Name:             "web-app",
			Comment:          "",
			Organization:     utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
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
		Organization: utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
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
		LastSpec:   utils.ObjectIdHex("688c7cde9da165ffad4b52e4"),
		DeploySpec: utils.ObjectIdHex("688c7cde9da165ffad4b52e4"),
		Hash:       "80309b44139a78378c9025d40535c73f52f9d71c",
	},
	{
		Id:           utils.ObjectIdHex("68b67d1aee12c08a1f39f88b"),
		Pod:          utils.ObjectIdHex("688bf358d978631566998ffc"),
		Organization: utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
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
		LastSpec:   utils.ObjectIdHex("688c7cde9da165ffad4b34f2"),
		DeploySpec: utils.ObjectIdHex("688c7cde9da165ffad4b34f2"),
		Hash:       "d0c176ab5dafce10956e0d3d5e1320b2e496acff",
	},
}

var Specs = []*spec.Spec{
	{
		Id:        utils.ObjectIdHex("688c7cde9da165ffad4b52e4"),
		Unit:      utils.ObjectIdHex("688c716d9da165ffad4b3682"),
		Index:     2,
		Hash:      "80309b44139a78378c9025d40535c73f52f9d71c",
		Timestamp: time.Now().Add(-5 * time.Minute),
		Data: "```yaml" + `
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
	},
	{
		Id:        utils.ObjectIdHex("68b67f44ee12c08a1f39fdbe"),
		Unit:      utils.ObjectIdHex("688c716d9da165ffad4b3682"),
		Index:     1,
		Hash:      "7e4a9f2c8b3d6h5j1k9m0n2p4q8r3s7t5v8w2x34",
		Timestamp: time.Now().Add(-5 * time.Hour),
		Data: "```yaml" + `
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

## Startup

` + "```shell {phase=reboot}" + `
systemctl start nginx
` + "```",
	},
	{
		Id:        utils.ObjectIdHex("688c7cde9da165ffad4b34f2"),
		Unit:      utils.ObjectIdHex("68b67d1aee12c08a1f39f88b"),
		Index:     2,
		Hash:      "d0c176ab5dafce10956e0d3d5e1320b2e496acff",
		Timestamp: time.Now().Add(-10 * time.Minute),
		Data: "```yaml" + `
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
	},
	{
		Id:        utils.ObjectIdHex("68b67cb1ee12c08a1f39f78b"),
		Unit:      utils.ObjectIdHex("68b67d1aee12c08a1f39f88b"),
		Index:     1,
		Hash:      "3c8f5a1d9e7b2k4m6n0p3q7r9s2t5v8w1x4y6z4d",
		Timestamp: time.Now().Add(-6 * time.Hour),
		Data: "```yaml" + `
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
	},
}

var SpecsNamed = []*spec.Named{
	{
		Id:        Specs[0].Id,
		Unit:      Specs[0].Unit,
		Index:     Specs[0].Index,
		Timestamp: Specs[0].Timestamp,
	},
	{
		Id:        Specs[1].Id,
		Unit:      Specs[1].Unit,
		Index:     Specs[1].Index,
		Timestamp: Specs[1].Timestamp,
	},
	{
		Id:        Specs[2].Id,
		Unit:      Specs[2].Unit,
		Index:     Specs[2].Index,
		Timestamp: Specs[2].Timestamp,
	},
	{
		Id:        Specs[3].Id,
		Unit:      Specs[3].Unit,
		Index:     Specs[3].Index,
		Timestamp: Specs[3].Timestamp,
	},
}

var DeploymentLogs = []string{
	"[2025-08-02 07:21:42] dnf install -y nginx\n",
	"[2025-08-02 07:21:43] Waiting for process with pid 1024 to finish.\n",
	"[2025-08-02 07:22:03] Last metadata expiration check: 0:00:02 ago on Sat 02 Aug 2025 07:22:01 AM UTC.\n",
	"[2025-08-02 07:22:04] Dependencies resolved.\n",
	"[2025-08-02 07:22:04] ================================================================================\n",
	"[2025-08-02 07:22:04]  Package             Arch    Version                   Repository          Size\n",
	"[2025-08-02 07:22:04] ================================================================================\n",
	"[2025-08-02 07:22:04] Installing:\n",
	"[2025-08-02 07:22:04]  nginx               x86_64  2:1.20.1-22.0.1.el9_6.3   ol9_appstream       49 k\n",
	"[2025-08-02 07:22:04] Installing dependencies:\n",
	"[2025-08-02 07:22:04]  nginx-core          x86_64  2:1.20.1-22.0.1.el9_6.3   ol9_appstream      589 k\n",
	"[2025-08-02 07:22:04]  nginx-filesystem    noarch  2:1.20.1-22.0.1.el9_6.3   ol9_appstream      9.6 k\n",
	"[2025-08-02 07:22:04]  oracle-logos-httpd  noarch  90.4-1.0.1.el9            ol9_baseos_latest   37 k\n",
	"[2025-08-02 07:22:04] Transaction Summary\n",
	"[2025-08-02 07:22:04] ================================================================================\n",
	"[2025-08-02 07:22:04] Install  4 Packages\n",
	"[2025-08-02 07:22:04] Total download size: 684 k\n",
	"[2025-08-02 07:22:04] Installed size: 1.8 M\n",
	"[2025-08-02 07:22:04] Downloading Packages:\n",
	"[2025-08-02 07:22:04] (1/4): nginx-1.20.1-22.0.1.el9_6.3.x86_64.rpm   757 kB/s |  49 kB     00:00\n",
	"[2025-08-02 07:22:04] (2/4): oracle-logos-httpd-90.4-1.0.1.el9.noarch 529 kB/s |  37 kB     00:00\n",
	"[2025-08-02 07:22:04] (3/4): nginx-filesystem-1.20.1-22.0.1.el9_6.3.n 540 kB/s | 9.6 kB     00:00\n",
	"[2025-08-02 07:22:04] (4/4): nginx-core-1.20.1-22.0.1.el9_6.3.x86_64. 5.7 MB/s | 589 kB     00:00\n",
	"[2025-08-02 07:22:04] --------------------------------------------------------------------------------\n",
	"[2025-08-02 07:22:04] Total                                           6.3 MB/s | 684 kB     00:00\n",
	"[2025-08-02 07:22:05] Running transaction check\n",
	"[2025-08-02 07:22:05] Transaction check succeeded.\n",
	"[2025-08-02 07:22:05] Running transaction test\n",
	"[2025-08-02 07:22:05] Transaction test succeeded.\n",
	"[2025-08-02 07:22:05] Running transaction\n",
	"[2025-08-02 07:22:05]   Preparing        :                                                        1/1\n",
	"[2025-08-02 07:22:05]   Running scriptlet: nginx-filesystem-2:1.20.1-22.0.1.el9_6.3.noarch        1/4\n",
	"[2025-08-02 07:22:05]   Installing       : nginx-filesystem-2:1.20.1-22.0.1.el9_6.3.noarch        1/4\n",
	"[2025-08-02 07:22:05]   Installing       : nginx-core-2:1.20.1-22.0.1.el9_6.3.x86_64              2/4\n",
	"[2025-08-02 07:22:05]   Installing       : oracle-logos-httpd-90.4-1.0.1.el9.noarch               3/4\n",
	"[2025-08-02 07:22:05]   Installing       : nginx-2:1.20.1-22.0.1.el9_6.3.x86_64                   4/4\n",
	"[2025-08-02 07:22:06]   Running scriptlet: nginx-2:1.20.1-22.0.1.el9_6.3.x86_64                   4/4\n",
	"[2025-08-02 07:22:06]   Verifying        : oracle-logos-httpd-90.4-1.0.1.el9.noarch               1/4\n",
	"[2025-08-02 07:22:06]   Verifying        : nginx-2:1.20.1-22.0.1.el9_6.3.x86_64                   2/4\n",
	"[2025-08-02 07:22:06]   Verifying        : nginx-core-2:1.20.1-22.0.1.el9_6.3.x86_64              3/4\n",
	"[2025-08-02 07:22:07]   Verifying        : nginx-filesystem-2:1.20.1-22.0.1.el9_6.3.noarch        4/4\n",
	"[2025-08-02 07:22:07] Installed:\n",
	"[2025-08-02 07:22:07]   nginx-2:1.20.1-22.0.1.el9_6.3.x86_64\n",
	"[2025-08-02 07:22:07]   nginx-core-2:1.20.1-22.0.1.el9_6.3.x86_64\n",
	"[2025-08-02 07:22:07]   nginx-filesystem-2:1.20.1-22.0.1.el9_6.3.noarch\n",
	"[2025-08-02 07:22:07]   oracle-logos-httpd-90.4-1.0.1.el9.noarch\n",
	"[2025-08-02 07:22:07] Complete!\n",
	"[2025-08-02 07:22:07] systemctl start nginx\n",
	"[2025-08-02 07:22:12] [INFO] ▶ agent: Queuing engine reload ◆ hash=2415297466 ◆ spec_len=618",
}

var Deployments = []*aggregate.Deployment{
	{
		Id:            utils.ObjectIdHex("651d8e7c4cf91e3b53d62d00"),
		Pod:           utils.ObjectIdHex("688bf358d978631566998ffc"),
		Unit:          utils.ObjectIdHex("688c716d9da165ffad4b3682"),
		Spec:          utils.ObjectIdHex("688c7cde9da165ffad4b52e4"),
		SpecOffset:    0,
		SpecIndex:     2,
		SpecTimestamp: time.Now(),
		Timestamp:     time.Now(),
		Tags:          []string{},
		Kind:          "instance",
		State:         "deployed",
		Action:        "",
		Status:        "healthy",
		Node:          utils.ObjectIdHex("689733b2a7a35eae0dbaea0a"),
		Instance:      utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a00"),
		InstanceData: &deployment.InstanceData{
			HostIps:     []string{"198.18.84.178"},
			PublicIps:   []string{"1.253.67.10"},
			PublicIps6:  []string{"2001:db8:85a3:4d2f:ac50:8355:bb57:e0f5"},
			PrivateIps:  []string{"10.196.3.18"},
			PrivateIps6: []string{"fd97:30bf:d456:a3bc:58d5:f529:66f2:36ef"},
		},
		ZoneName:            "us-west-1a",
		NodeName:            "pritunl-east0",
		InstanceName:        "web-app",
		InstanceRoles:       []string{"instance"},
		InstanceMemory:      2048,
		InstanceProcessors:  2,
		InstanceStatus:      "Running",
		InstanceUptime:      "5 days",
		InstanceState:       "running",
		InstanceAction:      "start",
		InstanceGuestStatus: "running",
		InstanceTimestamp:   time.Now(),
		InstanceHeartbeat:   time.Now(),
		InstanceMemoryUsage: 50.52,
		InstanceHugePages:   0,
		InstanceLoad1:       35.43,
		InstanceLoad5:       44.56,
		InstanceLoad15:      51.32,
	},
	{
		Id:            utils.ObjectIdHex("651d8e7c4cf91e3b53d62d01"),
		Pod:           utils.ObjectIdHex("688bf358d978631566998ffc"),
		Unit:          utils.ObjectIdHex("688c716d9da165ffad4b3682"),
		Spec:          utils.ObjectIdHex("688c7cde9da165ffad4b52e4"),
		SpecOffset:    0,
		SpecIndex:     2,
		SpecTimestamp: time.Now(),
		Timestamp:     time.Now(),
		Tags:          []string{},
		Kind:          "instance",
		State:         "deployed",
		Action:        "",
		Status:        "healthy",
		Node:          utils.ObjectIdHex("689733b2a7a35eae0dbaea0b"),
		Instance:      utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a01"),
		InstanceData: &deployment.InstanceData{
			HostIps:     []string{"198.18.84.82"},
			PublicIps:   []string{"1.253.67.103"},
			PublicIps6:  []string{"2001:db8:85a3:4d2f:5e1b:773a:2463:da58"},
			PrivateIps:  []string{"10.196.6.231"},
			PrivateIps6: []string{"fd97:30bf:d456:a3bc:bf1a:d5e4:56a2:4b27"},
		},
		ZoneName:            "us-west-1a",
		NodeName:            "pritunl-east1",
		InstanceName:        "web-app",
		InstanceRoles:       []string{"instance"},
		InstanceMemory:      2048,
		InstanceProcessors:  2,
		InstanceStatus:      "Running",
		InstanceUptime:      "5 days",
		InstanceState:       "running",
		InstanceAction:      "start",
		InstanceGuestStatus: "running",
		InstanceTimestamp:   time.Now(),
		InstanceHeartbeat:   time.Now(),
		InstanceMemoryUsage: 62.72,
		InstanceHugePages:   0,
		InstanceLoad1:       25.34,
		InstanceLoad5:       29.71,
		InstanceLoad15:      33.64,
	},
	{
		Id:            utils.ObjectIdHex("651d8e7c4cf91e3b53d62d02"),
		Pod:           utils.ObjectIdHex("688bf358d978631566998ffc"),
		Unit:          utils.ObjectIdHex("688c716d9da165ffad4b3682"),
		Spec:          utils.ObjectIdHex("688c7cde9da165ffad4b52e4"),
		SpecOffset:    0,
		SpecIndex:     2,
		SpecTimestamp: time.Now(),
		Timestamp:     time.Now(),
		Tags:          []string{},
		Kind:          "instance",
		State:         "deployed",
		Action:        "",
		Status:        "healthy",
		Node:          utils.ObjectIdHex("689733b2a7a35eae0dbaea0c"),
		Instance:      utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a02"),
		InstanceData: &deployment.InstanceData{
			HostIps:     []string{"198.18.84.79"},
			PublicIps:   []string{"1.253.67.148"},
			PublicIps6:  []string{"2001:db8:85a3:4d2f:27fe:0397:17fa:5d2e"},
			PrivateIps:  []string{"10.196.3.251"},
			PrivateIps6: []string{"fd97:30bf:d456:a3bc:3d61:c9f7:d2d7:8b9b"},
		},
		ZoneName:            "us-west-1a",
		NodeName:            "pritunl-east2",
		InstanceName:        "web-app",
		InstanceRoles:       []string{"instance"},
		InstanceMemory:      2048,
		InstanceProcessors:  2,
		InstanceStatus:      "Running",
		InstanceUptime:      "5 days",
		InstanceState:       "running",
		InstanceAction:      "start",
		InstanceGuestStatus: "running",
		InstanceTimestamp:   time.Now(),
		InstanceHeartbeat:   time.Now(),
		InstanceMemoryUsage: 56.29,
		InstanceHugePages:   0,
		InstanceLoad1:       50.22,
		InstanceLoad5:       58.12,
		InstanceLoad15:      66.57,
	},
	{
		Id:            utils.ObjectIdHex("651d8e7c4cf91e3b53d62d03"),
		Pod:           utils.ObjectIdHex("688bf358d978631566998ffc"),
		Unit:          utils.ObjectIdHex("688c716d9da165ffad4b3682"),
		Spec:          utils.ObjectIdHex("688c7cde9da165ffad4b52e4"),
		SpecOffset:    0,
		SpecIndex:     2,
		SpecTimestamp: time.Now(),
		Timestamp:     time.Now(),
		Tags:          []string{},
		Kind:          "instance",
		State:         "deployed",
		Action:        "",
		Status:        "healthy",
		Node:          utils.ObjectIdHex("689733b2a7a35eae0dbaea0d"),
		Instance:      utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a03"),
		InstanceData: &deployment.InstanceData{
			HostIps:     []string{"198.18.84.109"},
			PublicIps:   []string{"1.253.67.214"},
			PublicIps6:  []string{"2001:db8:85a3:4d2f:41b2:61e2:ad56:6cdc"},
			PrivateIps:  []string{"10.196.2.12"},
			PrivateIps6: []string{"fd97:30bf:d456:a3bc:c166:fabc:4223:a974"},
		},
		ZoneName:            "us-west-1a",
		NodeName:            "pritunl-east3",
		InstanceName:        "web-app",
		InstanceRoles:       []string{"instance"},
		InstanceMemory:      2048,
		InstanceProcessors:  2,
		InstanceStatus:      "Running",
		InstanceUptime:      "5 days",
		InstanceState:       "running",
		InstanceAction:      "start",
		InstanceGuestStatus: "running",
		InstanceTimestamp:   time.Now(),
		InstanceHeartbeat:   time.Now(),
		InstanceMemoryUsage: 32.21,
		InstanceHugePages:   0,
		InstanceLoad1:       54.49,
		InstanceLoad5:       59.36,
		InstanceLoad15:      64.21,
	},
	{
		Id:            utils.ObjectIdHex("651d8e7c4cf91e3b53d62d04"),
		Pod:           utils.ObjectIdHex("688bf358d978631566998ffc"),
		Unit:          utils.ObjectIdHex("688c716d9da165ffad4b3682"),
		Spec:          utils.ObjectIdHex("688c7cde9da165ffad4b52e4"),
		SpecOffset:    0,
		SpecIndex:     2,
		SpecTimestamp: time.Now(),
		Timestamp:     time.Now(),
		Tags:          []string{},
		Kind:          "instance",
		State:         "deployed",
		Action:        "",
		Status:        "healthy",
		Node:          utils.ObjectIdHex("689733b2a7a35eae0dbaea0e"),
		Instance:      utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a04"),
		InstanceData: &deployment.InstanceData{
			HostIps:     []string{"198.18.84.97"},
			PublicIps:   []string{"1.253.67.129"},
			PublicIps6:  []string{"2001:db8:85a3:4d2f:cb29:095b:a7f8:9a7e"},
			PrivateIps:  []string{"10.196.6.229"},
			PrivateIps6: []string{"fd97:30bf:d456:a3bc:b2f4:9b35:700e:0b9a"},
		},
		ZoneName:            "us-west-1a",
		NodeName:            "pritunl-east4",
		InstanceName:        "web-app",
		InstanceRoles:       []string{"instance"},
		InstanceMemory:      2048,
		InstanceProcessors:  2,
		InstanceStatus:      "Running",
		InstanceUptime:      "5 days",
		InstanceState:       "running",
		InstanceAction:      "start",
		InstanceGuestStatus: "running",
		InstanceTimestamp:   time.Now(),
		InstanceHeartbeat:   time.Now(),
		InstanceMemoryUsage: 33.96,
		InstanceHugePages:   0,
		InstanceLoad1:       14.25,
		InstanceLoad5:       18.58,
		InstanceLoad15:      24.81,
	},
	{
		Id:            utils.ObjectIdHex("651d8e7c4cf91e3b53d62d05"),
		Pod:           utils.ObjectIdHex("688bf358d978631566998ffc"),
		Unit:          utils.ObjectIdHex("688c716d9da165ffad4b3682"),
		Spec:          utils.ObjectIdHex("688c7cde9da165ffad4b52e4"),
		SpecOffset:    0,
		SpecIndex:     2,
		SpecTimestamp: time.Now(),
		Timestamp:     time.Now(),
		Tags:          []string{},
		Kind:          "instance",
		State:         "deployed",
		Action:        "",
		Status:        "healthy",
		Node:          utils.ObjectIdHex("689733b2a7a35eae0dbaea0f"),
		Instance:      utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a05"),
		InstanceData: &deployment.InstanceData{
			HostIps:     []string{"198.18.84.32"},
			PublicIps:   []string{"1.253.67.144"},
			PublicIps6:  []string{"2001:db8:85a3:4d2f:126f:552b:77d0:010e"},
			PrivateIps:  []string{"10.196.7.148"},
			PrivateIps6: []string{"fd97:30bf:d456:a3bc:c057:f8fa:ff43:a21a"},
		},
		ZoneName:            "us-west-1a",
		NodeName:            "pritunl-east5",
		InstanceName:        "web-app",
		InstanceRoles:       []string{"instance"},
		InstanceMemory:      2048,
		InstanceProcessors:  2,
		InstanceStatus:      "Running",
		InstanceUptime:      "5 days",
		InstanceState:       "running",
		InstanceAction:      "start",
		InstanceGuestStatus: "running",
		InstanceTimestamp:   time.Now(),
		InstanceHeartbeat:   time.Now(),
		InstanceMemoryUsage: 60.63,
		InstanceHugePages:   0,
		InstanceLoad1:       58.13,
		InstanceLoad5:       61.62,
		InstanceLoad15:      64.1,
	},
	{
		Id:            utils.ObjectIdHex("651d8e7c4cf91e3b53d62d06"),
		Pod:           utils.ObjectIdHex("688bf358d978631566998ffc"),
		Unit:          utils.ObjectIdHex("688c716d9da165ffad4b3682"),
		Spec:          utils.ObjectIdHex("688c7cde9da165ffad4b52e4"),
		SpecOffset:    0,
		SpecIndex:     2,
		SpecTimestamp: time.Now(),
		Timestamp:     time.Now(),
		Tags:          []string{},
		Kind:          "instance",
		State:         "deployed",
		Action:        "",
		Status:        "healthy",
		Node:          utils.ObjectIdHex("689733b2a7a35eae0dbaea0a"),
		Instance:      utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a06"),
		InstanceData: &deployment.InstanceData{
			HostIps:     []string{"198.18.84.87"},
			PublicIps:   []string{"1.253.67.73"},
			PublicIps6:  []string{"2001:db8:85a3:4d2f:aa28:1b64:c808:caeb"},
			PrivateIps:  []string{"10.196.4.215"},
			PrivateIps6: []string{"fd97:30bf:d456:a3bc:9797:d44a:0c7e:cb9e"},
		},
		ZoneName:            "us-west-1a",
		NodeName:            "pritunl-east0",
		InstanceName:        "web-app",
		InstanceRoles:       []string{"instance"},
		InstanceMemory:      2048,
		InstanceProcessors:  2,
		InstanceStatus:      "Running",
		InstanceUptime:      "5 days",
		InstanceState:       "running",
		InstanceAction:      "start",
		InstanceGuestStatus: "running",
		InstanceTimestamp:   time.Now(),
		InstanceHeartbeat:   time.Now(),
		InstanceMemoryUsage: 48.03,
		InstanceHugePages:   0,
		InstanceLoad1:       53.4,
		InstanceLoad5:       60.79,
		InstanceLoad15:      68.75,
	},
	{
		Id:            utils.ObjectIdHex("651d8e7c4cf91e3b53d62d07"),
		Pod:           utils.ObjectIdHex("688bf358d978631566998ffc"),
		Unit:          utils.ObjectIdHex("688c716d9da165ffad4b3682"),
		Spec:          utils.ObjectIdHex("688c7cde9da165ffad4b52e4"),
		SpecOffset:    0,
		SpecIndex:     2,
		SpecTimestamp: time.Now(),
		Timestamp:     time.Now(),
		Tags:          []string{},
		Kind:          "instance",
		State:         "deployed",
		Action:        "",
		Status:        "healthy",
		Node:          utils.ObjectIdHex("689733b2a7a35eae0dbaea0b"),
		Instance:      utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a07"),
		InstanceData: &deployment.InstanceData{
			HostIps:     []string{"198.18.84.151"},
			PublicIps:   []string{"1.253.67.65"},
			PublicIps6:  []string{"2001:db8:85a3:4d2f:bf64:91d6:4050:eac0"},
			PrivateIps:  []string{"10.196.5.97"},
			PrivateIps6: []string{"fd97:30bf:d456:a3bc:0dd4:8931:8c28:5465"},
		},
		ZoneName:            "us-west-1a",
		NodeName:            "pritunl-east1",
		InstanceName:        "web-app",
		InstanceRoles:       []string{"instance"},
		InstanceMemory:      2048,
		InstanceProcessors:  2,
		InstanceStatus:      "Running",
		InstanceUptime:      "5 days",
		InstanceState:       "running",
		InstanceAction:      "start",
		InstanceGuestStatus: "running",
		InstanceTimestamp:   time.Now(),
		InstanceHeartbeat:   time.Now(),
		InstanceMemoryUsage: 78.17,
		InstanceHugePages:   0,
		InstanceLoad1:       34.25,
		InstanceLoad5:       40.18,
		InstanceLoad15:      46.49,
	},
	{
		Id:            utils.ObjectIdHex("651d8e7c4cf91e3b53d62d08"),
		Pod:           utils.ObjectIdHex("688bf358d978631566998ffc"),
		Unit:          utils.ObjectIdHex("688c716d9da165ffad4b3682"),
		Spec:          utils.ObjectIdHex("688c7cde9da165ffad4b52e4"),
		SpecOffset:    0,
		SpecIndex:     2,
		SpecTimestamp: time.Now(),
		Timestamp:     time.Now(),
		Tags:          []string{},
		Kind:          "instance",
		State:         "deployed",
		Action:        "",
		Status:        "healthy",
		Node:          utils.ObjectIdHex("689733b2a7a35eae0dbaea0c"),
		Instance:      utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a08"),
		InstanceData: &deployment.InstanceData{
			HostIps:     []string{"198.18.84.224"},
			PublicIps:   []string{"1.253.67.211"},
			PublicIps6:  []string{"2001:db8:85a3:4d2f:f5f3:98f7:82b5:ee87"},
			PrivateIps:  []string{"10.196.3.34"},
			PrivateIps6: []string{"fd97:30bf:d456:a3bc:8d49:241d:4dd1:4663"},
		},
		ZoneName:            "us-west-1a",
		NodeName:            "pritunl-east2",
		InstanceName:        "web-app",
		InstanceRoles:       []string{"instance"},
		InstanceMemory:      2048,
		InstanceProcessors:  2,
		InstanceStatus:      "Running",
		InstanceUptime:      "5 days",
		InstanceState:       "running",
		InstanceAction:      "start",
		InstanceGuestStatus: "running",
		InstanceTimestamp:   time.Now(),
		InstanceHeartbeat:   time.Now(),
		InstanceMemoryUsage: 53.78,
		InstanceHugePages:   0,
		InstanceLoad1:       35.24,
		InstanceLoad5:       38.31,
		InstanceLoad15:      43.56,
	},
	{
		Id:            utils.ObjectIdHex("651d8e7c4cf91e3b53d62d09"),
		Pod:           utils.ObjectIdHex("688bf358d978631566998ffc"),
		Unit:          utils.ObjectIdHex("688c716d9da165ffad4b3682"),
		Spec:          utils.ObjectIdHex("688c7cde9da165ffad4b52e4"),
		SpecOffset:    0,
		SpecIndex:     2,
		SpecTimestamp: time.Now(),
		Timestamp:     time.Now(),
		Tags:          []string{},
		Kind:          "instance",
		State:         "deployed",
		Action:        "",
		Status:        "healthy",
		Node:          utils.ObjectIdHex("689733b2a7a35eae0dbaea0d"),
		Instance:      utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a09"),
		InstanceData: &deployment.InstanceData{
			HostIps:     []string{"198.18.84.9"},
			PublicIps:   []string{"1.253.67.121"},
			PublicIps6:  []string{"2001:db8:85a3:4d2f:6e3a:29b0:639f:49d4"},
			PrivateIps:  []string{"10.196.2.187"},
			PrivateIps6: []string{"fd97:30bf:d456:a3bc:ddb4:e207:cc09:d1e6"},
		},
		ZoneName:            "us-west-1a",
		NodeName:            "pritunl-east3",
		InstanceName:        "web-app",
		InstanceRoles:       []string{"instance"},
		InstanceMemory:      2048,
		InstanceProcessors:  2,
		InstanceStatus:      "Running",
		InstanceUptime:      "5 days",
		InstanceState:       "running",
		InstanceAction:      "start",
		InstanceGuestStatus: "running",
		InstanceTimestamp:   time.Now(),
		InstanceHeartbeat:   time.Now(),
		InstanceMemoryUsage: 42.23,
		InstanceHugePages:   0,
		InstanceLoad1:       56.09,
		InstanceLoad5:       57.62,
		InstanceLoad15:      65.05,
	},
	{
		Id:            utils.ObjectIdHex("651d8e7c4cf91e3b53d62d0a"),
		Pod:           utils.ObjectIdHex("688bf358d978631566998ffc"),
		Unit:          utils.ObjectIdHex("68b67d1aee12c08a1f39f88b"),
		Spec:          utils.ObjectIdHex("688c7cde9da165ffad4b34f2"),
		SpecOffset:    0,
		SpecIndex:     2,
		SpecTimestamp: time.Now(),
		Timestamp:     time.Now(),
		Tags:          []string{},
		Kind:          "instance",
		State:         "deployed",
		Action:        "",
		Status:        "healthy",
		Node:          utils.ObjectIdHex("689733b2a7a35eae0dbaea0e"),
		Instance:      utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a0a"),
		InstanceData: &deployment.InstanceData{
			HostIps:     []string{"198.18.84.61"},
			PublicIps:   []string{"1.253.67.205"},
			PublicIps6:  []string{"2001:db8:85a3:4d2f:d943:7ff9:dfdc:4d68"},
			PrivateIps:  []string{"10.196.3.219"},
			PrivateIps6: []string{"fd97:30bf:d456:a3bc:7ba9:bca3:5217:b534"},
		},
		ZoneName:            "us-west-1a",
		NodeName:            "pritunl-east4",
		InstanceName:        "database",
		InstanceRoles:       []string{"instance"},
		InstanceMemory:      8192,
		InstanceProcessors:  4,
		InstanceStatus:      "Running",
		InstanceUptime:      "5 days",
		InstanceState:       "running",
		InstanceAction:      "start",
		InstanceGuestStatus: "running",
		InstanceTimestamp:   time.Now(),
		InstanceHeartbeat:   time.Now(),
		InstanceMemoryUsage: 62.84,
		InstanceHugePages:   0,
		InstanceLoad1:       40.76,
		InstanceLoad5:       46.28,
		InstanceLoad15:      48.62,
	},
	{
		Id:            utils.ObjectIdHex("651d8e7c4cf91e3b53d62d0b"),
		Pod:           utils.ObjectIdHex("688bf358d978631566998ffc"),
		Unit:          utils.ObjectIdHex("68b67d1aee12c08a1f39f88b"),
		Spec:          utils.ObjectIdHex("688c7cde9da165ffad4b34f2"),
		SpecOffset:    0,
		SpecIndex:     2,
		SpecTimestamp: time.Now(),
		Timestamp:     time.Now(),
		Tags:          []string{},
		Kind:          "instance",
		State:         "deployed",
		Action:        "",
		Status:        "healthy",
		Node:          utils.ObjectIdHex("689733b2a7a35eae0dbaea0f"),
		Instance:      utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a0b"),
		InstanceData: &deployment.InstanceData{
			HostIps:     []string{"198.18.84.221"},
			PublicIps:   []string{"1.253.67.155"},
			PublicIps6:  []string{"2001:db8:85a3:4d2f:3e3e:0d9d:8669:2c89"},
			PrivateIps:  []string{"10.196.8.253"},
			PrivateIps6: []string{"fd97:30bf:d456:a3bc:d0ff:8b42:1d9b:92fd"},
		},
		ZoneName:            "us-west-1a",
		NodeName:            "pritunl-east5",
		InstanceName:        "database",
		InstanceRoles:       []string{"instance"},
		InstanceMemory:      8192,
		InstanceProcessors:  4,
		InstanceStatus:      "Running",
		InstanceUptime:      "5 days",
		InstanceState:       "running",
		InstanceAction:      "start",
		InstanceGuestStatus: "running",
		InstanceTimestamp:   time.Now(),
		InstanceHeartbeat:   time.Now(),
		InstanceMemoryUsage: 66.15,
		InstanceHugePages:   0,
		InstanceLoad1:       28.1,
		InstanceLoad5:       32.45,
		InstanceLoad15:      40.43,
	},
	{
		Id:            utils.ObjectIdHex("651d8e7c4cf91e3b53d62d0c"),
		Pod:           utils.ObjectIdHex("688bf358d978631566998ffc"),
		Unit:          utils.ObjectIdHex("68b67d1aee12c08a1f39f88b"),
		Spec:          utils.ObjectIdHex("688c7cde9da165ffad4b34f2"),
		SpecOffset:    0,
		SpecIndex:     2,
		SpecTimestamp: time.Now(),
		Timestamp:     time.Now(),
		Tags:          []string{},
		Kind:          "instance",
		State:         "deployed",
		Action:        "",
		Status:        "healthy",
		Node:          utils.ObjectIdHex("689733b2a7a35eae0dbaea0a"),
		Instance:      utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a0c"),
		InstanceData: &deployment.InstanceData{
			HostIps:     []string{"198.18.84.99"},
			PublicIps:   []string{"1.253.67.59"},
			PublicIps6:  []string{"2001:db8:85a3:4d2f:74b0:8661:53c1:1d5b"},
			PrivateIps:  []string{"10.196.2.110"},
			PrivateIps6: []string{"fd97:30bf:d456:a3bc:e7b4:670b:acf5:dfb4"},
		},
		ZoneName:            "us-west-1a",
		NodeName:            "pritunl-east0",
		InstanceName:        "database",
		InstanceRoles:       []string{"instance"},
		InstanceMemory:      8192,
		InstanceProcessors:  4,
		InstanceStatus:      "Running",
		InstanceUptime:      "5 days",
		InstanceState:       "running",
		InstanceAction:      "start",
		InstanceGuestStatus: "running",
		InstanceTimestamp:   time.Now(),
		InstanceHeartbeat:   time.Now(),
		InstanceMemoryUsage: 56.45,
		InstanceHugePages:   0,
		InstanceLoad1:       27.42,
		InstanceLoad5:       29.06,
		InstanceLoad15:      36.64,
	},
	{
		Id:            utils.ObjectIdHex("651d8e7c4cf91e3b53d62d0d"),
		Pod:           utils.ObjectIdHex("688bf358d978631566998ffc"),
		Unit:          utils.ObjectIdHex("68b67d1aee12c08a1f39f88b"),
		Spec:          utils.ObjectIdHex("688c7cde9da165ffad4b34f2"),
		SpecOffset:    0,
		SpecIndex:     2,
		SpecTimestamp: time.Now(),
		Timestamp:     time.Now(),
		Tags:          []string{},
		Kind:          "instance",
		State:         "deployed",
		Action:        "",
		Status:        "healthy",
		Node:          utils.ObjectIdHex("689733b2a7a35eae0dbaea0b"),
		Instance:      utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a0d"),
		InstanceData: &deployment.InstanceData{
			HostIps:     []string{"198.18.84.208"},
			PublicIps:   []string{"1.253.67.39"},
			PublicIps6:  []string{"2001:db8:85a3:4d2f:4b40:60d1:ed30:0b06"},
			PrivateIps:  []string{"10.196.6.194"},
			PrivateIps6: []string{"fd97:30bf:d456:a3bc:5220:ac62:3c7c:7291"},
		},
		ZoneName:            "us-west-1a",
		NodeName:            "pritunl-east1",
		InstanceName:        "database",
		InstanceRoles:       []string{"instance"},
		InstanceMemory:      8192,
		InstanceProcessors:  4,
		InstanceStatus:      "Running",
		InstanceUptime:      "5 days",
		InstanceState:       "running",
		InstanceAction:      "start",
		InstanceGuestStatus: "running",
		InstanceTimestamp:   time.Now(),
		InstanceHeartbeat:   time.Now(),
		InstanceMemoryUsage: 54.62,
		InstanceHugePages:   0,
		InstanceLoad1:       54.37,
		InstanceLoad5:       57.22,
		InstanceLoad15:      63.01,
	},
	{
		Id:            utils.ObjectIdHex("651d8e7c4cf91e3b53d62d0e"),
		Pod:           utils.ObjectIdHex("688bf358d978631566998ffc"),
		Unit:          utils.ObjectIdHex("68b67d1aee12c08a1f39f88b"),
		Spec:          utils.ObjectIdHex("688c7cde9da165ffad4b34f2"),
		SpecOffset:    0,
		SpecIndex:     2,
		SpecTimestamp: time.Now(),
		Timestamp:     time.Now(),
		Tags:          []string{},
		Kind:          "instance",
		State:         "deployed",
		Action:        "",
		Status:        "healthy",
		Node:          utils.ObjectIdHex("689733b2a7a35eae0dbaea0c"),
		Instance:      utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a0e"),
		InstanceData: &deployment.InstanceData{
			HostIps:     []string{"198.18.84.210"},
			PublicIps:   []string{"1.253.67.19"},
			PublicIps6:  []string{"2001:db8:85a3:4d2f:be8e:3013:d6f6:5396"},
			PrivateIps:  []string{"10.196.6.223"},
			PrivateIps6: []string{"fd97:30bf:d456:a3bc:c924:41b8:22f3:5b3f"},
		},
		ZoneName:            "us-west-1a",
		NodeName:            "pritunl-east2",
		InstanceName:        "database",
		InstanceRoles:       []string{"instance"},
		InstanceMemory:      8192,
		InstanceProcessors:  4,
		InstanceStatus:      "Running",
		InstanceUptime:      "5 days",
		InstanceState:       "running",
		InstanceAction:      "start",
		InstanceGuestStatus: "running",
		InstanceTimestamp:   time.Now(),
		InstanceHeartbeat:   time.Now(),
		InstanceMemoryUsage: 43.89,
		InstanceHugePages:   0,
		InstanceLoad1:       46.21,
		InstanceLoad5:       53.63,
		InstanceLoad15:      62.25,
	},
	{
		Id:            utils.ObjectIdHex("651d8e7c4cf91e3b53d62d0f"),
		Pod:           utils.ObjectIdHex("688bf358d978631566998ffc"),
		Unit:          utils.ObjectIdHex("68b67d1aee12c08a1f39f88b"),
		Spec:          utils.ObjectIdHex("688c7cde9da165ffad4b34f2"),
		SpecOffset:    0,
		SpecIndex:     2,
		SpecTimestamp: time.Now(),
		Timestamp:     time.Now(),
		Tags:          []string{},
		Kind:          "instance",
		State:         "deployed",
		Action:        "",
		Status:        "healthy",
		Node:          utils.ObjectIdHex("689733b2a7a35eae0dbaea0d"),
		Instance:      utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a0f"),
		InstanceData: &deployment.InstanceData{
			HostIps:     []string{"198.18.84.50"},
			PublicIps:   []string{"1.253.67.60"},
			PublicIps6:  []string{"2001:db8:85a3:4d2f:fff2:877d:227c:1c4a"},
			PrivateIps:  []string{"10.196.7.165"},
			PrivateIps6: []string{"fd97:30bf:d456:a3bc:ae17:b804:32c5:956c"},
		},
		ZoneName:            "us-west-1a",
		NodeName:            "pritunl-east3",
		InstanceName:        "database",
		InstanceRoles:       []string{"instance"},
		InstanceMemory:      8192,
		InstanceProcessors:  4,
		InstanceStatus:      "Running",
		InstanceUptime:      "5 days",
		InstanceState:       "running",
		InstanceAction:      "start",
		InstanceGuestStatus: "running",
		InstanceTimestamp:   time.Now(),
		InstanceHeartbeat:   time.Now(),
		InstanceMemoryUsage: 41.89,
		InstanceHugePages:   0,
		InstanceLoad1:       24.24,
		InstanceLoad5:       25.81,
		InstanceLoad15:      30.23,
	},
}
