package demo

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/advisory"
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vulnerability"
)

var advisoryInstancesInfo = []*aggregate.AdvisoryInstanceInfo{
	{
		Id:              utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a08"),
		Name:            "web-app",
		Status:          "Running",
		Timestamp:       time.Now(),
		Uptime:          "5 days 11 hours 34 mins",
		PublicIps:       []string{"1.253.67.197"},
		PublicIps6:      []string{"2001:db8:85a3:4d2f:f5f3:98f7:82b5:ee87"},
		PrivateIps:      []string{"10.196.4.2"},
		PrivateIps6:     []string{"fd97:30bf:d456:a3bc:8d49:241d:4dd1:4663"},
		CloudPublicIps:  []string{},
		CloudPublicIps6: []string{},
	},
	{
		Id:              utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a09"),
		Name:            "web-app",
		Status:          "Running",
		Timestamp:       time.Now(),
		Uptime:          "5 days 11 hours 34 mins",
		PublicIps:       []string{"1.253.67.63"},
		PublicIps6:      []string{"2001:db8:85a3:4d2f:6e3a:29b0:639f:49d4"},
		PrivateIps:      []string{"10.196.6.18"},
		PrivateIps6:     []string{"fd97:30bf:d456:a3bc:ddb4:e207:cc09:d1e6"},
		CloudPublicIps:  []string{},
		CloudPublicIps6: []string{},
	},
	{
		Id:              utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a10"),
		Name:            "search",
		Status:          "Running",
		Timestamp:       time.Now(),
		Uptime:          "5 days 11 hours 34 mins",
		PublicIps:       []string{"1.253.67.118"},
		PublicIps6:      []string{"2001:db8:85a3:4d2f:b5e8:fca8:79fa:a3b4"},
		PrivateIps:      []string{"10.196.4.224"},
		PrivateIps6:     []string{"fd97:30bf:d456:a3bc:83e0:cd9a:85c4:38da"},
		CloudPublicIps:  []string{},
		CloudPublicIps6: []string{},
	},
	{
		Id:              utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a11"),
		Name:            "search",
		Status:          "Running",
		Timestamp:       time.Now(),
		Uptime:          "5 days 11 hours 34 mins",
		PublicIps:       []string{"1.253.67.230"},
		PublicIps6:      []string{"2001:db8:85a3:4d2f:508e:01c4:330a:418c"},
		PrivateIps:      []string{"10.196.4.33"},
		PrivateIps6:     []string{"fd97:30bf:d456:a3bc:d6a1:6424:48d3:bdfc"},
		CloudPublicIps:  []string{},
		CloudPublicIps6: []string{},
	},
}

var advisoryNodesInfo = []*aggregate.AdvisoryNodeInfo{
	{
		Id:         utils.ObjectIdHex("689733b2a7a35eae0dbaea0e"),
		Name:       "pritunl-east4",
		Timestamp:  time.Now(),
		PublicIps:  []string{"10.253.67.94"},
		PublicIps6: []string{"2001:db8:85a3:4d2f:6f15:d8c3:2a94:e7b1"},
		PrivateIps: []string{"10.8.0.14"},
	},
	{
		Id:         utils.ObjectIdHex("689733b2a7a35eae0dbaea0f"),
		Name:       "pritunl-east5",
		Timestamp:  time.Now(),
		PublicIps:  []string{"10.253.67.95"},
		PublicIps6: []string{"2001:db8:85a3:4d2f:8d2c:1fa7:b653:9e48"},
		PrivateIps: []string{"10.8.0.15"},
	},
}

var Advisories = []*aggregate.AdvisoryAggregate{
	{
		Advisory: advisory.Advisory{
			Id:           utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56b00"),
			Organization: utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
			Reference:    "ALSA-2026:1472",
			Type:         advisory.RedHat,
			Updated:      time.Date(2026, 5, 22, 23, 38, 56, 0, time.UTC),
			Severity:     "important",
			Description:  "OpenSSL is a toolkit that implements the Secure Sockets Layer (SSL) and Transport Layer Security (TLS) protocols, as well as a full-strength general-purpose cryptography library.\n\nSecurity Fix(es):\n\n* openssl: OpenSSL: Arbitrary code execution or denial of service through crafted PKCS#12 file (CVE-2025-11187)\n* openssl: OpenSSL: Remote code execution or Denial of Service via oversized Initialization Vector in CMS parsing (CVE-2025-15467)\n* openssl: OpenSSL: Denial of Service via NULL pointer dereference in QUIC protocol handling (CVE-2025-15468)\n* openssl: OpenSSL: Data integrity bypass in `openssl dgst` command due to silent truncation (CVE-2025-15469)\n* openssl: OpenSSL: Denial of Service due to excessive memory allocation in TLS 1.3 certificate compression (CVE-2025-66199)\n* openssl: OpenSSL: Denial of Service due to out-of-bounds write in BIO filter (CVE-2025-68160)\n* openssl: OpenSSL: Information disclosure and data tampering via specific low-level OCB encryption/decryption calls (CVE-2025-69418)\n* openssl: OpenSSL: Arbitrary code execution due to out-of-bounds write in PKCS#12 processing (CVE-2025-69419)\n* openssl: OpenSSL: Denial of Service via malformed PKCS#12 file processing (CVE-2025-69421)\n* openssl: OpenSSL: Denial of Service via malformed TimeStamp Response (CVE-2025-69420)\n* openssl: OpenSSL: Denial of Service due to type confusion in PKCS#12 file processing (CVE-2026-22795)\n* openssl: OpenSSL: Denial of Service via type confusion in PKCS#7 signature verification (CVE-2026-22796)\n\n\nFor more details about the security issue(s), including the impact, a CVSS score, acknowledgments, and other related information, refer to the CVE page(s) listed in the References section.",
			Score:        advisory.Critical,
			Packages: []string{
				"openssl-1:3.5.1-7.el10_1.alma.1.x86_64",
				"openssl-libs-1:3.5.1-7.el10_1.alma.1.x86_64",
			},
			Vuxmls: []string{},
			Vulnerabilities: []*vulnerability.Vulnerability{
				{
					Id:        "rh:CVE-2025-11187",
					Timestamp: time.Date(2026, 5, 22, 23, 38, 45, 38000000, time.UTC),
					Severity:  vulnerability.Medium,
					Score:     6.1,
				},
				{
					Id:        "rh:CVE-2025-15467",
					Timestamp: time.Date(2026, 5, 22, 23, 38, 46, 44000000, time.UTC),
					Severity:  vulnerability.High,
					Score:     9.8,
				},
				{
					Id:        "rh:CVE-2025-15468",
					Timestamp: time.Date(2026, 5, 22, 23, 38, 47, 39000000, time.UTC),
					Severity:  vulnerability.Low,
					Score:     5.9,
				},
				{
					Id:        "rh:CVE-2025-15469",
					Timestamp: time.Date(2026, 5, 22, 23, 38, 48, 40000000, time.UTC),
					Severity:  vulnerability.Low,
					Score:     5.5,
				},
				{
					Id:        "rh:CVE-2025-66199",
					Timestamp: time.Date(2026, 5, 22, 23, 38, 49, 41000000, time.UTC),
					Severity:  vulnerability.Low,
					Score:     5.9,
				},
				{
					Id:        "rh:CVE-2025-68160",
					Timestamp: time.Date(2026, 5, 22, 23, 38, 50, 46000000, time.UTC),
					Severity:  vulnerability.Low,
					Score:     4.7,
				},
				{
					Id:        "rh:CVE-2025-69418",
					Timestamp: time.Date(2026, 5, 22, 23, 38, 51, 74000000, time.UTC),
					Severity:  vulnerability.Low,
					Score:     4.0,
				},
				{
					Id:        "rh:CVE-2025-69419",
					Timestamp: time.Date(2026, 5, 22, 23, 38, 52, 74000000, time.UTC),
					Severity:  vulnerability.Medium,
					Score:     7.4,
				},
				{
					Id:        "rh:CVE-2025-69420",
					Timestamp: time.Date(2026, 5, 22, 23, 38, 53, 90000000, time.UTC),
					Severity:  vulnerability.Low,
					Score:     5.9,
				},
				{
					Id:        "rh:CVE-2025-69421",
					Timestamp: time.Date(2026, 5, 22, 23, 38, 54, 65000000, time.UTC),
					Severity:  vulnerability.Low,
					Score:     6.5,
				},
				{
					Id:        "rh:CVE-2026-22795",
					Timestamp: time.Date(2026, 5, 22, 23, 38, 55, 169000000, time.UTC),
					Severity:  vulnerability.Low,
					Score:     5.5,
				},
				{
					Id:        "rh:CVE-2026-22796",
					Timestamp: time.Date(2026, 5, 22, 23, 38, 56, 78000000, time.UTC),
					Severity:  vulnerability.Low,
					Score:     5.9,
				},
			},
			Instances: []bson.ObjectID{
				utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a08"),
				utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a09"),
				utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a10"),
				utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a11"),
			},
			Nodes: []bson.ObjectID{
				utils.ObjectIdHex("689733b2a7a35eae0dbaea0e"),
				utils.ObjectIdHex("689733b2a7a35eae0dbaea0f"),
			},
			DismissedResources: []bson.ObjectID{},
		},
		InstancesInfo: advisoryInstancesInfo,
		NodesInfo:     advisoryNodesInfo,
	},
	{
		Advisory: advisory.Advisory{
			Id:           utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56b01"),
			Organization: utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
			Reference:    "ALSA-2026:13380",
			Type:         advisory.RedHat,
			Updated:      time.Date(2026, 5, 22, 23, 38, 43, 0, time.UTC),
			Severity:     "important",
			Description:  "OpenSSH is an SSH protocol implementation supported by a number of Linux, UNIX, and similar operating systems. It includes the core files necessary for both the OpenSSH client and server.\n\nSecurity Fix(es):\n\n* OpenSSH: OpenSSH: Privilege escalation via scp legacy protocol when not preserving file mode (CVE-2026-35385)\n* OpenSSH: OpenSSH: Security bypass via mishandling of authorized_keys principals option (CVE-2026-35414)\n* OpenSSH: OpenSSH: Information disclosure due to unintended cryptographic algorithm usage (CVE-2026-35387)\n* OpenSSH: OpenSSH: Low integrity impact from unconfirmed proxy-mode multiplexing sessions (CVE-2026-35388)\n* OpenSSH: OpenSSH: Arbitrary command execution via shell metacharacters in username (CVE-2026-35386)\n\n\nFor more details about the security issue(s), including the impact, a CVSS score, acknowledgments, and other related information, refer to the CVE page(s) listed in the References section.",
			Score:        advisory.High,
			Packages: []string{
				"openssh-9.9p1-14.el10_1.alma.1.x86_64",
				"openssh-clients-9.9p1-14.el10_1.alma.1.x86_64",
				"openssh-server-9.9p1-14.el10_1.alma.1.x86_64",
			},
			Vuxmls: []string{},
			Vulnerabilities: []*vulnerability.Vulnerability{
				{
					Id:        "rh:CVE-2026-35385",
					Timestamp: time.Date(2026, 5, 22, 23, 38, 39, 110000000, time.UTC),
					Severity:  vulnerability.High,
					Score:     7.5,
				},
				{
					Id:        "rh:CVE-2026-35386",
					Timestamp: time.Date(2026, 5, 22, 23, 38, 40, 86000000, time.UTC),
					Severity:  vulnerability.Low,
					Score:     3.6,
				},
				{
					Id:        "rh:CVE-2026-35387",
					Timestamp: time.Date(2026, 5, 22, 23, 38, 41, 42000000, time.UTC),
					Severity:  vulnerability.Low,
					Score:     3.1,
				},
				{
					Id:        "rh:CVE-2026-35388",
					Timestamp: time.Date(2026, 5, 22, 23, 38, 42, 77000000, time.UTC),
					Severity:  vulnerability.Low,
					Score:     2.2,
				},
				{
					Id:        "rh:CVE-2026-35414",
					Timestamp: time.Date(2026, 5, 22, 23, 38, 43, 56000000, time.UTC),
					Severity:  vulnerability.Medium,
					Score:     4.8,
				},
			},
			Instances: []bson.ObjectID{
				utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a08"),
				utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a09"),
				utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a10"),
				utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a11"),
			},
			Nodes: []bson.ObjectID{
				utils.ObjectIdHex("689733b2a7a35eae0dbaea0e"),
				utils.ObjectIdHex("689733b2a7a35eae0dbaea0f"),
			},
			DismissedResources: []bson.ObjectID{},
		},
		InstancesInfo: advisoryInstancesInfo,
		NodesInfo:     advisoryNodesInfo,
	},
	{
		Advisory: advisory.Advisory{
			Id:           utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56b02"),
			Organization: utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
			Reference:    "ALSA-2026:4012",
			Type:         advisory.RedHat,
			Updated:      time.Date(2026, 5, 22, 23, 39, 48, 0, time.UTC),
			Severity:     "moderate",
			Description:  "The kernel packages contain the Linux kernel, the core of any Linux operating system.\n\nSecurity Fix(es):\n\n* kernel: Linux kernel: Use-after-free in device mapper due to race condition in zone reporting (CVE-2025-38141)\n* kernel: Linux kernel io_uring: Local privilege escalation, information disclosure, or denial of service via use-after-free (CVE-2025-38106)\n* kernel: drm/xe: Make dma-fences compliant with the safe access rules (CVE-2025-38703)\n* kernel: Linux kernel: Denial of Service via out-of-bounds read in USB configuration parsing (CVE-2025-39760)\n* kernel: HID: intel-thc-hid: intel-thc: Fix incorrect pointer arithmetic in I2C regs save (CVE-2025-39818)\n* kernel: Kernel: Use-after-free in GPIO character device allows privilege escalation or denial of service (CVE-2025-40249)\n* kernel: ipv6: BUG() in pskb_expand_head() as part of calipso_skbuff_setattr() (CVE-2025-71085)\n* kernel: macvlan: fix possible UAF in macvlan_forward_source() (CVE-2026-23001)\n* kernel: Linux kernel: Denial of Service due to a deadlock in hugetlb folio migration (CVE-2026-23097)\n* kernel: Linux kernel: Information disclosure in efivarfs via incorrect error propagation (CVE-2026-23156)\n\n\nFor more details about the security issue(s), including the impact, a CVSS score, acknowledgments, and other related information, refer to the CVE page(s) listed in the References section.",
			Score:        advisory.Medium,
			Packages: []string{
				"kernel-6.12.0-124.43.1.el10_1.x86_64",
				"kernel-core-6.12.0-124.43.1.el10_1.x86_64",
				"kernel-modules-6.12.0-124.43.1.el10_1.x86_64",
				"kernel-modules-core-6.12.0-124.43.1.el10_1.x86_64",
				"kernel-modules-extra-6.12.0-124.43.1.el10_1.x86_64",
				"kernel-modules-extra-matched-6.12.0-124.43.1.el10_1.x86_64",
				"kernel-tools-6.12.0-124.43.1.el10_1.x86_64",
				"kernel-tools-libs-6.12.0-124.43.1.el10_1.x86_64",
				"python3-perf-6.12.0-124.43.1.el10_1.x86_64",
			},
			Vuxmls: []string{},
			Vulnerabilities: []*vulnerability.Vulnerability{
				{
					Id:        "rh:CVE-2025-38106",
					Timestamp: time.Date(2026, 5, 22, 23, 39, 39, 108000000, time.UTC),
					Severity:  vulnerability.Medium,
					Score:     7.1,
				},
				{
					Id:        "rh:CVE-2025-38141",
					Timestamp: time.Date(2026, 5, 22, 23, 39, 40, 101000000, time.UTC),
					Severity:  vulnerability.Medium,
					Score:     7.0,
				},
				{
					Id:        "rh:CVE-2025-38703",
					Timestamp: time.Date(2026, 5, 22, 23, 39, 41, 220000000, time.UTC),
					Severity:  vulnerability.Medium,
					Score:     7.0,
				},
				{
					Id:        "rh:CVE-2025-39760",
					Timestamp: time.Date(2026, 5, 22, 23, 39, 42, 153000000, time.UTC),
					Severity:  vulnerability.Medium,
					Score:     7.1,
				},
				{
					Id:        "rh:CVE-2025-39818",
					Timestamp: time.Date(2026, 5, 22, 23, 39, 43, 95000000, time.UTC),
					Severity:  vulnerability.Medium,
					Score:     5.5,
				},
				{
					Id:        "rh:CVE-2025-40249",
					Timestamp: time.Date(2026, 5, 22, 23, 39, 44, 137000000, time.UTC),
					Severity:  vulnerability.Medium,
					Score:     7.1,
				},
				{
					Id:        "rh:CVE-2025-71085",
					Timestamp: time.Date(2026, 5, 22, 23, 39, 45, 159000000, time.UTC),
					Severity:  vulnerability.Medium,
					Score:     7.5,
				},
				{
					Id:        "rh:CVE-2026-23001",
					Timestamp: time.Date(2026, 5, 22, 23, 39, 46, 126000000, time.UTC),
					Severity:  vulnerability.Medium,
					Score:     7.8,
				},
				{
					Id:        "rh:CVE-2026-23097",
					Timestamp: time.Date(2026, 5, 22, 23, 39, 47, 80000000, time.UTC),
					Severity:  vulnerability.Medium,
					Score:     7.3,
				},
				{
					Id:        "rh:CVE-2026-23156",
					Timestamp: time.Date(2026, 5, 22, 23, 39, 48, 104000000, time.UTC),
					Severity:  vulnerability.Medium,
					Score:     7.3,
				},
			},
			Instances: []bson.ObjectID{
				utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a08"),
				utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a09"),
				utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a10"),
				utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56a11"),
			},
			Nodes: []bson.ObjectID{
				utils.ObjectIdHex("689733b2a7a35eae0dbaea0e"),
				utils.ObjectIdHex("689733b2a7a35eae0dbaea0f"),
			},
			DismissedResources: []bson.ObjectID{},
		},
		InstancesInfo: advisoryInstancesInfo,
		NodesInfo:     advisoryNodesInfo,
	},
}
