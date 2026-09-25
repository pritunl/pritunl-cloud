package demo

import (
	"sort"
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/advisory"
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vulnerability"
)

var advisoryInstancesInfo = getAdvisoryInstancesInfo()

func getAdvisoryInstancesInfo() (infos []*aggregate.AdvisoryInstanceInfo) {
	infos = []*aggregate.AdvisoryInstanceInfo{}

	for _, inst := range Instances {
		if inst.Name != "web-app" && inst.Name != "search" {
			continue
		}

		infos = append(infos, &aggregate.AdvisoryInstanceInfo{
			Id:              inst.Id,
			Name:            inst.Name,
			Status:          inst.Status,
			Timestamp:       inst.Timestamp,
			Uptime:          inst.Uptime,
			PublicIps:       inst.PublicIps,
			PublicIps6:      inst.PublicIps6,
			PrivateIps:      inst.PrivateIps,
			PrivateIps6:     inst.PrivateIps6,
			CloudPublicIps:  []string{},
			CloudPublicIps6: []string{},
		})
	}

	return
}

var advisoryNodesInfo = getAdvisoryNodesInfo()

func getAdvisoryNodesInfo() (infos []*aggregate.AdvisoryNodeInfo) {
	infos = []*aggregate.AdvisoryNodeInfo{}

	for _, nde := range Nodes {
		privateIps := []string{}
		for _, privateIp := range nde.PrivateIps {
			privateIps = append(privateIps, privateIp)
		}
		sort.Strings(privateIps)

		infos = append(infos, &aggregate.AdvisoryNodeInfo{
			Id:         nde.Id,
			Name:       nde.Name,
			Timestamp:  nde.Timestamp,
			PublicIps:  nde.PublicIps,
			PublicIps6: nde.PublicIps6,
			PrivateIps: privateIps,
		})
	}

	return
}

var Advisories = []*advisory.Advisory{
	{
		Id:           utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56b00"),
		Organization: utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
		Reference:    "ALSA-2026:25191",
		Type:         advisory.RedHat,
		Updated:      time.Date(2026, 9, 24, 7, 52, 0, 278000000, time.UTC),
		Severity:     "critical",
		Description:  "The kernel packages contain the Linux kernel, the core of any Linux operating system.\n\nSecurity Fix(es):\n\n* kernel: Linux kernel: Use-after-free in bonding driver leads to denial of service (CVE-2026-31419)\n* kernel: Linux kernel: Denial of Service in erofs filesystem (CVE-2026-31467)\n* kernel: can: raw: fix ro->uniq use-after-free in raw_rcv() (CVE-2026-31532)\n* kernel: ALSA: 6fire: fix use-after-free on disconnect (CVE-2026-31581)\n* kernel: ip6_tunnel: clear skb2->cb[] in ip4ip6_err() (CVE-2026-43037)\n* kernel: ipv6: rpl: reserve mac_len headroom when recompressed SRH grows (CVE-2026-43501)\n* kernel: selinux: fix overlayfs mmap() and mprotect() access checks (CVE-2026-46054)\n\n\nFor more details about the security issue(s), including the impact, a CVSS score, acknowledgments, and other related information, refer to the CVE page(s) listed in the References section.",
		Score:        advisory.Critical,
		Packages: []string{
			"kernel-6.12.0-211.22.1.el10_2.x86_64",
			"kernel-core-6.12.0-211.22.1.el10_2.x86_64",
			"kernel-modules-6.12.0-211.22.1.el10_2.x86_64",
			"kernel-modules-core-6.12.0-211.22.1.el10_2.x86_64",
			"kernel-modules-extra-6.12.0-211.22.1.el10_2.x86_64",
			"kernel-modules-extra-matched-6.12.0-211.22.1.el10_2.x86_64",
			"kernel-tools-6.12.0-211.22.1.el10_2.x86_64",
			"kernel-tools-libs-6.12.0-211.22.1.el10_2.x86_64",
			"python3-perf-6.12.0-211.22.1.el10_2.x86_64",
		},
		Vuxmls: []string{},
		Vulnerabilities: []string{
			"CVE-2026-31419",
			"CVE-2026-31467",
			"CVE-2026-31532",
			"CVE-2026-31581",
			"CVE-2026-43037",
			"CVE-2026-43501",
			"CVE-2026-46054",
		},
		Complete:      true,
		InstanceCount: len(advisoryInstancesInfo),
		NodeCount:     len(advisoryNodesInfo),
	},
	{
		Id:           utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56b01"),
		Organization: utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
		Reference:    "ALSA-2026:67314",
		Type:         advisory.RedHat,
		Updated:      time.Date(2026, 9, 24, 8, 15, 29, 705000000, time.UTC),
		Severity:     "important",
		Description:  "nginx is a web and proxy server supporting HTTP and other protocols, with a focus on high concurrency, performance, and low memory usage.\n\nSecurity Fix(es):\n\n* nginx: NGINX: Arbitrary code execution via crafted HTTP requests (CVE-2026-42533)\n\n\nFor more details about the security issue(s), including the impact, a CVSS score, acknowledgments, and other related information, refer to the CVE page(s) listed in the References section.",
		Score:        advisory.High,
		Packages: []string{
			"nginx-2:1.26.3-6.el10_2.7.x86_64",
			"nginx-core-2:1.26.3-6.el10_2.7.x86_64",
			"nginx-filesystem-2:1.26.3-6.el10_2.7.noarch",
		},
		Vuxmls: []string{},
		Vulnerabilities: []string{
			"CVE-2026-42533",
		},
		Complete:      true,
		InstanceCount: len(advisoryInstancesInfo),
		NodeCount:     len(advisoryNodesInfo),
	},
	{
		Id:           utils.ObjectIdHex("651d8e7c4cf9e2e3e4d56b02"),
		Organization: utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
		Reference:    "ALSA-2026:66355",
		Type:         advisory.RedHat,
		Updated:      time.Date(2026, 9, 24, 7, 52, 0, 278000000, time.UTC),
		Severity:     "important",
		Description:  "The kernel packages contain the Linux kernel, the core of any Linux operating system.\n\nSecurity Fix(es):\n\n* kernel: net: bridge: stop fast-leave after deleting a port group (CVE-2026-74480)\n\n\nFor more details about the security issue(s), including the impact, a CVSS score, acknowledgments, and other related information, refer to the CVE page(s) listed in the References section.",
		Score:        advisory.Medium,
		Packages: []string{
			"kernel-6.12.0-211.54.1.el10_2.x86_64",
			"kernel-core-6.12.0-211.54.1.el10_2.x86_64",
			"kernel-headers-6.12.0-211.54.1.el10_2.x86_64",
			"kernel-modules-6.12.0-211.54.1.el10_2.x86_64",
			"kernel-modules-core-6.12.0-211.54.1.el10_2.x86_64",
			"kernel-modules-extra-6.12.0-211.54.1.el10_2.x86_64",
			"kernel-modules-extra-matched-6.12.0-211.54.1.el10_2.x86_64",
			"kernel-tools-6.12.0-211.54.1.el10_2.x86_64",
			"kernel-tools-libs-6.12.0-211.54.1.el10_2.x86_64",
			"python3-perf-6.12.0-211.54.1.el10_2.x86_64",
		},
		Vuxmls: []string{},
		Vulnerabilities: []string{
			"CVE-2026-74480",
		},
		Complete:      true,
		InstanceCount: len(advisoryInstancesInfo),
		NodeCount:     len(advisoryNodesInfo),
	},
}

var advisoryVulnerabilities = []*vulnerability.Vulnerability{
	{
		Id:              "CVE-2026-31419",
		Source:          vulnerability.RedHat,
		Timestamp:       time.Date(2026, 9, 24, 5, 34, 32, 347000000, time.UTC),
		Published:       time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC),
		Status:          vulnerability.Analyzed,
		Description:     "In the Linux kernel, the following vulnerability has been resolved:\nnet: bonding: fix use-after-free in bond_xmit_broadcast()\nbond_xmit_broadcast() reuses the original skb for the last slave\n(determined by bond_is_last_slave()) and clones it for others.\nConcurrent slave enslave/release can mutate the slave list during\nRCU-protected iteration, changing which slave is \"last\" mid-loop.\nThis causes the original skb to be double-consumed (double-freed).\nReplace the racy bond_is_last_slave() check with a simple index\ncomparison (i + 1 == slaves_count) against the pre-snapshot slave\ncount taken via READ_ONCE() before the loop.  This preserves the\nzero-copy optimization for the last slave while making the \"last\"\ndetermination stable against concurrent list mutations.\nThe UAF can trigger the following crash:\n==================================================================\nBUG: KASAN: slab-use-after-free in skb_clone\nRead of size 8 at addr ffff888100ef8d40 by task exploit/147\nCPU: 1 UID: 0 PID: 147 Comm: exploit Not tainted 7.0.0-rc3+ #4 PREEMPTLAZY\nCall Trace:\n<TASK>\ndump_stack_lvl (lib/dump_stack.c:123)\nprint_report (mm/kasan/report.c:379 mm/kasan/report.c:482)\nkasan_report (mm/kasan/report.c:597)\nskb_clone (include/linux/skbuff.h:1724 include/linux/skbuff.h:1792 include/linux/skbuff.h:3396 net/core/skbuff.c:2108)\nbond_xmit_broadcast (drivers/net/bonding/bond_main.c:5334)\nbond_start_xmit (drivers/net/bonding/bond_main.c:5567 drivers/net/bonding/bond_main.c:5593)\ndev_hard_start_xmit (include/linux/netdevice.h:5325 include/linux/netdevice.h:5334 net/core/dev.c:3871 net/core/dev.c:3887)\n__dev_queue_xmit (include/linux/netdevice.h:3601 net/core/dev.c:4838)\nip6_finish_output2 (include/net/neighbour.h:540 include/net/neighbour.h:554 net/ipv6/ip6_output.c:136)\nip6_finish_output (net/ipv6/ip6_output.c:208 net/ipv6/ip6_output.c:219)\nip6_output (net/ipv6/ip6_output.c:250)\nip6_send_skb (net/ipv6/ip6_output.c:1985)\nudp_v6_send_skb (net/ipv6/udp.c:1442)\nudpv6_sendmsg (net/ipv6/udp.c:1733)\n__sys_sendto (net/socket.c:730 net/socket.c:742 net/socket.c:2206)\n__x64_sys_sendto (net/socket.c:2209)\ndo_syscall_64 (arch/x86/entry/syscall_64.c:63 arch/x86/entry/syscall_64.c:94)\nentry_SYSCALL_64_after_hwframe (arch/x86/entry/entry_64.S:130)\n</TASK>\nAllocated by task 147:\nFreed by task 147:\nThe buggy address belongs to the object at ffff888100ef8c80\nwhich belongs to the cache skbuff_head_cache of size 224\nThe buggy address is located 192 bytes inside of\nfreed 224-byte region [ffff888100ef8c80, ffff888100ef8d60)\nMemory state around the buggy address:\nffff888100ef8c00: fb fb fb fb fc fc fc fc fc fc fc fc fc fc fc fc\nffff888100ef8c80: fa fb fb fb fb fb fb fb fb fb fb fb fb fb fb fb\n>ffff888100ef8d00: fb fb fb fb fb fb fb fb fb fb fb fb fc fc fc fc\n^\nffff888100ef8d80: fc fc fc fc fc fc fc fc fa fb fb fb fb fb fb fb\nffff888100ef8e00: fb fb fb fb fb fb fb fb fb fb fb fb fb fb fb fb\n==================================================================\n\nA flaw was found in the Linux kernel's bonding driver. A local attacker with low privileges could exploit a use-after-free vulnerability in the `bond_xmit_broadcast()` function. This occurs due to a race condition during concurrent slave enslave/release operations, which can lead to the original socket buffer (skb) being double-freed. Successful exploitation of this flaw can result in a system crash, leading to a denial of service.",
		Statement:       "This is an Important impact flaw affecting the Linux kernel's bonding driver in Red Hat Enterprise Linux 6, 8.8 and later, 9.2 and later, and 10, as well as Red Hat In-Vehicle OS 2.0. A local attacker with low privileges could trigger a use-after-free vulnerability, leading to a system crash and denial of service. Red Hat Enterprise Linux 7, 8.2, 8.4, 8.6, and 9.0 are not affected as the vulnerable code is not present.",
		Score:           7,
		Severity:        vulnerability.High,
		Vector:          vulnerability.Local,
		Complexity:      vulnerability.High,
		Privileges:      vulnerability.Low,
		Interaction:     vulnerability.None,
		Scope:           vulnerability.Unchanged,
		Confidentiality: vulnerability.High,
		Integrity:       vulnerability.High,
		Availability:    vulnerability.High,
		Analysis: &vulnerability.Analysis{
			Updated:   time.Date(2026, 9, 24, 5, 34, 32, 347000000, time.UTC),
			Timestamp: time.Date(2026, 9, 24, 5, 21, 36, 938000000, time.UTC),
			Model:     "claude-opus-5-5",
			Summary:   "This flaw is reachable in principle on a default server: an unprivileged local user can create a user and network namespace, autoload the bonding module, build a broadcast-mode bond with dummy slaves, and race packet sends against slave enslave/release. A stable-list backport post from 2026 says exactly this and reports a KASAN use-after-free reproduced on 6.1.182. No public privilege-escalation exploit or in-the-wild use has been reported. The upstream fix (commit 2884bf72fb8f, merged into net at the end of March 2026) changes one line in drivers/net/bonding/bond_main.c, in bond_xmit_broadcast(). The old code chose the 'last' slave with bond_is_last_slave() while walking an RCU-protected slave list. The last slave gets the original skb and the others get clones. If a slave is added or removed during the walk, which slave counts as 'last' can change mid-loop, so the original skb is handed to two transmitters: a double free / use-after-free in the skbuff_head_cache slab. The fix compares the loop index against a slave count snapshotted with READ_ONCE() before the loop. The trigger needs CAP_NET_ADMIN over a network namespace, which an unprivileged user gets through an unprivileged user namespace (enabled by default on RHEL/AlmaLinux). The bonding.ko module autoloads on demand through the rtnl-link-bond alias when the netlink request creates a bond link, so a loaded-module check will only see it after the attempt unless the host already uses bonding. Full impact depends on winning a narrow race between sendmsg and slave release and then grooming the freed skb object. The realistic result is a kernel crash (DoS); a skilled attacker could possibly turn the double free into local root, but nobody has shown that publicly. Red Hat lists RHEL 6, 8.8 and later, 9.2 and later, and 10 as affected, and RHEL 7, 8.2, 8.4, 8.6 and 9.0 as not affected. Ubuntu has shipped fixes in several USNs, and Debian is affected too. Stable backports (6.1, 6.6, 6.12) went through revert/re-apply rounds because of conflicts with the broadcast_neighbor patch, so check distribution kernel changelogs rather than assume a fix is in. Containers: the default Docker/Podman seccomp profiles block unshare(CLONE_NEWUSER) and containers get no CAP_NET_ADMIN, so a default container cannot reach this. Only containers that are privileged, have CAP_NET_ADMIN, or run with a relaxed seccomp profile that allows nested user namespaces can.",
			Impacted:  "* bonding.ko (Ethernet channel bonding driver): broadcast-mode transmit path double-frees the skb when slaves change during a send\n* Unprivileged user namespaces: give ordinary users the CAP_NET_ADMIN they need to create a bond and slaves in their own netns\n* Containers allowed to create user namespaces, or granted CAP_NET_ADMIN or privileged mode: can trigger it against the shared host kernel\n* Hosts running bonded NICs in mode=broadcast (balance-rr/802.3ad hosts still have the module loaded, but the attacker creates their own broadcast bond anyway)",
			Score:     7,
			Universal: true,
			Processes: []string{},
			Modules: []string{
				"bonding",
			},
			Ports: []*vulnerability.Port{},
			Evidence: []*vulnerability.Evidence{
				{
					Type: "fix_commit",
					Url:  "https://git.kernel.org/netdev/net/c/2884bf72fb8f",
					Note: "Upstream fix: bond_xmit_broadcast() now compares i + 1 against a slave count snapshotted with READ_ONCE() instead of calling the racy bond_is_last_slave()",
				},
				{
					Type: "other",
					Url:  "https://ratatoskr.run/netdev/2026/03/11351556/t",
					Note: "netdev thread for PATCH net v3 with the diff and the patchwork bot confirming it was applied",
				},
				{
					Type: "exploitation",
					Url:  "https://ratatoskr.run/stable/2026/08/17435092/t",
					Note: "6.1.y backport post: an unprivileged user can create a broadcast bond with dummy slaves in a user+net namespace; KASAN UAF reproduced on v6.1.182",
				},
				{
					Type: "other",
					Url:  "https://ratatoskr.run/stable/2026/06/17140875/t",
					Note: "6.12 stable series that reverts and re-applies the fix alongside the broadcast_neighbor dependency; shows backport churn",
				},
				{
					Type: "advisory",
					Url:  "https://ratatoskr.run/linux-cve-announce/2026/04/8048293",
					Note: "linux-cve-announce CVE assignment",
				},
				{
					Type: "advisory",
					Url:  "https://access.redhat.com/security/cve/cve-2026-31419",
					Note: "Red Hat: RHEL 6, 8.8+, 9.2+ and 10 affected; RHEL 7, 8.2-8.6 and 9.0 not affected",
				},
				{
					Type: "advisory",
					Url:  "https://ubuntu.com/security/CVE-2026-31419",
					Note: "Ubuntu CVE tracker; fixed through multiple USNs (e.g. USN-8277, USN-8305)",
				},
				{
					Type: "advisory",
					Url:  "https://bugzilla.redhat.com/show_bug.cgi?id=2457829",
					Note: "Red Hat bugzilla entry",
				},
			},
			Boundaries: []string{
				"local_dos",
				"user_to_root",
				"container_to_host",
			},
			Mitigation: "* Block the module where bonding is not used: add `install bonding /bin/false` to /etc/modprobe.d/ (do not do this on hosts that use NIC bonding)\n* Turn off unprivileged user namespaces: `sysctl -w user.max_user_namespaces=0` (RHEL/AlmaLinux), or on Ubuntu `kernel.apparmor_restrict_unprivileged_userns=1` / `kernel.unprivileged_userns_clone=0`\n* Keep the default container seccomp profile, which blocks unshare/clone with CLONE_NEWUSER, and do not grant CAP_NET_ADMIN or --privileged to untrusted containers",
		},
		Sync: vulnerability.Complete,
	},
	{
		Id:              "CVE-2026-31467",
		Source:          vulnerability.RedHat,
		Timestamp:       time.Date(2026, 9, 24, 5, 21, 58, 69000000, time.UTC),
		Published:       time.Date(2026, 4, 22, 0, 0, 0, 0, time.UTC),
		Status:          vulnerability.Analyzed,
		Description:     "In the Linux kernel, the following vulnerability has been resolved:\nerofs: add GFP_NOIO in the bio completion if needed\nThe bio completion path in the process context (e.g. dm-verity)\nwill directly call into decompression rather than trigger another\nworkqueue context for minimal scheduling latencies, which can\nthen call vm_map_ram() with GFP_KERNEL.\nDue to insufficient memory, vm_map_ram() may generate memory\nswapping I/O, which can cause submit_bio_wait to deadlock\nin some scenarios.\nTrimmed down the call stack, as follows:\nf2fs_submit_read_io\nsubmit_bio                      //bio_list is initialized.\nmmc_blk_mq_recovery\nz_erofs_endio\nvm_map_ram\n__pte_alloc_kernel\n__alloc_pages_direct_reclaim\nshrink_folio_list\n__swap_writepage\nsubmit_bio_wait  //bio_list is non-NULL, hang!!!\nUse memalloc_noio_{save,restore}() to wrap up this path.\n\nA flaw was found in the Linux kernel's erofs filesystem. A remote attacker can exploit this vulnerability without requiring any privileges. This issue occurs when insufficient memory during a memory mapping operation (vm_map_ram()) in the bio completion path leads to a deadlock, causing a Denial of Service (DoS).",
		Statement:       "",
		Score:           5.5,
		Severity:        vulnerability.Medium,
		Vector:          vulnerability.Local,
		Complexity:      vulnerability.Low,
		Privileges:      vulnerability.Low,
		Interaction:     vulnerability.None,
		Scope:           vulnerability.Unchanged,
		Confidentiality: vulnerability.None,
		Integrity:       vulnerability.None,
		Availability:    vulnerability.High,
		Analysis: &vulnerability.Analysis{
			Updated:   time.Date(2026, 9, 24, 5, 21, 58, 69000000, time.UTC),
			Timestamp: time.Date(2026, 9, 24, 5, 21, 31, 365000000, time.UTC),
			Model:     "claude-opus-5-5",
			Summary:   "Nobody has reported exploiting this flaw, and it is hard to trigger on a typical server. It is a rare deadlock that needs several things at once: a compressed EROFS filesystem already mounted by root, block I/O that completes in process context (the patch cites dm-verity, and the reported stack shows eMMC recovery on an Android/Amlogic device), heavy memory pressure, and active swap. The upstream fix (commit c23df30915f83e7257c8625b690a1cece94142a0, 'erofs: add GFP_NOIO in the bio completion if needed', by Jiucheng Xu of Amlogic, posted March 2026 and backported to stable branches including 5.15.y and 6.12.y) changes one path. When z_erofs_endio decompresses data directly inside the bio completion, the fix wraps that work in memalloc_noio_save()/memalloc_noio_restore(). Without it, vm_map_ram() allocates with GFP_KERNEL, and direct reclaim can start swap writeback through submit_bio_wait() while the current bio_list is still in use, which hangs the task. The only result is a hang or availability loss on the I/O path. There is no memory corruption, privilege gain or information leak. The vendor text claiming a 'remote attacker without privileges' is template wording and is not supported by the fix. Exposure: the erofs module ships as a loadable module in RHEL/AlmaLinux 9/10 (used by composefs, and Fedora/RHEL configs enable EROFS compression) and in Ubuntu kernels. However, EROFS is block-based and not FS_USERNS_MOUNT, so mounting it needs CAP_SYS_ADMIN in the initial namespace. An unprivileged user or container cannot mount a crafted image or autoload the module through a user namespace. At most, a local user could add memory pressure while reading files from an EROFS mount that an administrator created, hoping to hit the deadlock. Composefs metadata images are usually uncompressed, so the vulnerable decompression path generally does not run there. It matters mainly for compressed EROFS images, such as some container snapshotters or OS images, on dm-verity or similar setups where bio completion runs in process context. The boundary crossed is local DoS only. Distribution kernels older than the stable backport that include EROFS compression support are affected. I could not read the Ubuntu and Red Hat per-release fix status directly; only their CVE entries were confirmed to exist.",
			Impacted:  "* erofs kernel module: compressed (LZ4/LZMA/DEFLATE) images read via decompression in the bio completion can deadlock under memory pressure\n* dm-verity-protected EROFS images: process-context bio completion is what enables the direct decompression path\n* Container image stores that use compressed EROFS (containerd erofs snapshotter, Nydus, composefs with compressed images): readers can hang\n* Hosts with swap enabled: swap writeback during reclaim is what causes the hang",
			Score:     2.5,
			Universal: false,
			Processes: []string{},
			Modules: []string{
				"erofs",
			},
			Ports: []*vulnerability.Port{},
			Evidence: []*vulnerability.Evidence{
				{
					Type: "other",
					Url:  "https://ratatoskr.run/b4-sent/2026/03/3452815",
					Note: "PATCH v2 'erofs: add GFP_NOIO in the bio completion if needed' (Amlogic), wraps z_erofs_endio decompression in memalloc_noio_save/restore; upstream commit c23df30915f83e7257c8625b690a1cece94142a0",
				},
				{
					Type: "other",
					Url:  "https://ratatoskr.run/stable/2026/03/7140376/t",
					Note: "Stable backport request for 5.15.y and 6.12.y citing upstream commit c23df30915f8",
				},
				{
					Type: "other",
					Url:  "https://ratatoskr.run/linux-block/2026/03/3434420/t",
					Note: "linux-block discussion of the hang; trigger stack involves mmc_blk_mq_recovery -> z_erofs_endio -> vm_map_ram, with in_atomic-dependent direct decompression kept for latency",
				},
				{
					Type: "other",
					Url:  "https://www.mail-archive.com/kernel@lists.fedoraproject.org/msg17351.html",
					Note: "Fedora/RHEL kernel config enabling EROFS compression options for composefs use, showing the module is shipped in RHEL-family kernels",
				},
				{
					Type: "other",
					Url:  "https://www.man7.org/linux/man-pages/man7/user_namespaces.7.html",
					Note: "Block-based filesystems can only be mounted with CAP_SYS_ADMIN in the initial user namespace; EROFS is not user-namespace mountable",
				},
				{
					Type: "advisory",
					Url:  "https://ubuntu.com/security/CVE-2026-31467",
					Note: "Ubuntu CVE tracker entry",
				},
				{
					Type: "advisory",
					Url:  "https://app.opencve.io/cve/CVE-2026-31467",
					Note: "Aggregated record: CWE-667, Red Hat Moderate, local attack vector",
				},
			},
			Boundaries: []string{
				"local_dos",
			},
			Mitigation: "* If EROFS is not used (check for composefs or EROFS-based container snapshotters first), block the module with `install erofs /bin/false` in /etc/modprobe.d/erofs.conf\n* Use uncompressed EROFS images (mkfs.erofs without -z) so the z_erofs decompression path is never taken\n* On hosts that must mount compressed EROFS on dm-verity, running without swap (`swapoff -a`) removes the swap-writeback reclaim step that deadlocks",
		},
		Sync: vulnerability.Complete,
	},
	{
		Id:              "CVE-2026-31532",
		Source:          vulnerability.RedHat,
		Timestamp:       time.Date(2026, 9, 24, 5, 42, 5, 130000000, time.UTC),
		Published:       time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC),
		Status:          vulnerability.Analyzed,
		Description:     "In the Linux kernel, the following vulnerability has been resolved:\ncan: raw: fix ro->uniq use-after-free in raw_rcv()\nraw_release() unregisters raw CAN receive filters via can_rx_unregister(),\nbut receiver deletion is deferred with call_rcu(). This leaves a window\nwhere raw_rcv() may still be running in an RCU read-side critical section\nafter raw_release() frees ro->uniq, leading to a use-after-free of the\npercpu uniq storage.\nMove free_percpu(ro->uniq) out of raw_release() and into a raw-specific\nsocket destructor. can_rx_unregister() takes an extra reference to the\nsocket and only drops it from the RCU callback, so freeing uniq from\nsk_destruct ensures the percpu area is not released until the relevant\ncallbacks have drained.\n[mkl: applied manually]\n\nA flaw was found in the Linux kernel's Controller Area Network (CAN) raw socket implementation. A use-after-free vulnerability can occur due to a timing window during the unregistration of CAN receive filters, allowing a freed memory region to be accessed. This could lead to system instability or a denial of service (DoS).",
		Statement:       "A use-after-free flaw in the Linux kernel CAN raw socket implementation can occur when raw_release() frees the per-CPU ro->uniq storage before RCU-deferred receive callbacks have fully drained. A local attacker able to create raw CAN sockets and trigger CAN receive activity could race socket teardown with raw_rcv() and cause a kernel crash. The likely impact is denial of service, while privilege escalation would require additional, unproven control over per-CPU memory reuse.\nThe issue appears more consistent with an RCU lifetime/race-condition bug leading primarily to denial of service, as exploitation depends on a narrow teardown race involving a percpu object rather than a generic reclaimable slab object, with no demonstrated privilege-escalation path or obvious controlled overwrite primitive in the relatively niche CAN raw socket subsystem.",
		Score:           7.1,
		Severity:        vulnerability.Medium,
		Vector:          vulnerability.Local,
		Complexity:      vulnerability.Low,
		Privileges:      vulnerability.Low,
		Interaction:     vulnerability.None,
		Scope:           vulnerability.Unchanged,
		Confidentiality: vulnerability.High,
		Integrity:       vulnerability.None,
		Availability:    vulnerability.High,
		Analysis: &vulnerability.Analysis{
			Updated:   time.Date(2026, 9, 24, 5, 42, 5, 130000000, time.UTC),
			Timestamp: time.Date(2026, 9, 24, 5, 22, 13, 987000000, time.UTC),
			Model:     "claude-opus-5-5",
			Summary:   "This flaw can be reached in principle from an unprivileged local account on a default server that allows unprivileged user namespaces, but it depends on winning a narrow RCU race. As of this writing I found no reports of exploitation in the wild and no public privilege-escalation exploit; only the finder's write-up describing how the bug was validated. The bug is in net/can/raw.c and dates to v4.1 (commit 514ac99c64b2, which added the per-CPU ro->uniq duplicate-filtering state). raw_release() unregisters the socket's CAN filters with can_rx_unregister(), but the receiver is only removed later through call_rcu(). raw_release() then freed ro->uniq right away, so a raw_rcv() still running in softirq/RCU context could write to freed per-CPU memory (the skb pointer, skbcnt and join_rx_count fields). The upstream fix, a535a9217ca3 (backported to stable, e.g. 7201a531b9a5 and 572f0bf536eb), moves free_percpu(ro->uniq) into a new raw_sock_destruct() socket destructor. That destructor only runs after the RCU callback drops the socket reference. Trigger conditions: CAN_RAW sockets need no capability, and the can and can_raw modules autoload on socket(AF_CAN, SOCK_RAW, CAN_RAW). The attacker also needs a CAN interface to generate receive traffic. If the host has no physical CAN adapter, that means creating a vcan or vxcan device, which needs CAP_NET_ADMIN. An unprivileged user gets that capability by creating a user plus network namespace, which is the default on RHEL 8/9/10 and Ubuntu 22.04 and older (Ubuntu 23.10+ restricts it with AppArmor unless that is bypassed). The attacker then races close() of a bound raw socket against frames being delivered to it. Impact: a use-after-free write into the per-CPU allocator area from softirq context. The realistic result is a kernel crash or memory corruption (local DoS). Escalating to root would require controlling reuse of per-CPU chunks, which has not been shown and is much harder than typical slab use-after-free exploitation. The CVSS confidentiality-high rating is not supported by the fix; this is a stale write, not a read. For containers: Docker and Podman's default seccomp profile blocks unshare(CLONE_NEWUSER), so a default container can only trigger the bug if a CAN interface was passed in. Kubernetes pods that run seccomp-Unconfined (the default when no profile is set) on a host with user namespaces enabled can reach it. Every supported RHEL (8, 9, 10) and Ubuntu (20.04 to 24.04+) kernel carries the vulnerable code until patched. CAN is built as modules (can, can_raw, vcan); on some RHEL builds the CAN modules ship in kernel-modules-extra, in which case they are not installed by default. The modules normally load only at attack time, so a loaded-module check will not flag an unpatched host before an attack.",
			Impacted:  "* CAN raw sockets (can_raw module): per-CPU duplicate-filter state is freed while the receive path may still write to it\n* Unprivileged user namespaces: let local users or unconfined pods create a vcan/vxcan interface to drive the receive path\n* vcan/vxcan virtual CAN drivers: supply the frame traffic needed to win the race when no physical CAN hardware exists\n* Kubernetes pods running seccomp-Unconfined: can create namespaces and CAN interfaces and reach the bug\n* Physical CAN adapters (industrial/automotive gateways): hosts with real CAN interfaces are reachable without namespaces",
			Score:     6.5,
			Universal: true,
			Processes: []string{},
			Modules: []string{
				"can",
				"can_raw",
				"vcan",
				"vxcan",
			},
			Ports: []*vulnerability.Port{},
			Evidence: []*vulnerability.Evidence{
				{
					Type: "fix_commit",
					Url:  "https://git.kernel.org/stable/c/7201a531b9a5ed892bfda5ded9194ef622de8ffa",
					Note: "Stable backport of upstream a535a9217ca3: moves free_percpu(ro->uniq) from raw_release() into a new raw_sock_destruct() in net/can/raw.c",
				},
				{
					Type: "fix_commit",
					Url:  "https://git.kernel.org/stable/c/572f0bf536ebc14f6e7da3d21a85cf076de8358e",
					Note: "Second stable backport of the same fix",
				},
				{
					Type: "other",
					Url:  "https://ratatoskr.run/stable/2026/04/7143577",
					Note: "Archived copy of the netdev/linux-can patch 'can: raw: fix ro->uniq use-after-free in raw_rcv()', Fixes: 514ac99c64b2 (v4.1+), acked by Oliver Hartkopp, applied by Marc Kleine-Budde",
				},
				{
					Type: "other",
					Url:  "https://bynar.io/blog/discovery-validation-in-the-linux-kernel-part-1-can-use-after-free-race",
					Note: "Finder's write-up: CAN_RAW needs no CAP_NET_RAW; needs a CAN/vcan interface; creating vcan needs CAP_NET_ADMIN, so unprivileged user namespaces are required; Ubuntu 22.04 reachable by default",
				},
				{
					Type: "other",
					Url:  "https://edera.dev/stories/user-namespaces-are-not-a-security-boundary",
					Note: "Shows vcan and vxcan devices can be created from an unprivileged container using a user+net namespace, with the modules autoloading",
				},
				{
					Type: "advisory",
					Url:  "https://access.redhat.com/security/cve/cve-2026-31532",
					Note: "Red Hat advisory; vendor judges impact primarily DoS with privilege escalation unproven",
				},
				{
					Type: "advisory",
					Url:  "https://bugzilla.redhat.com/show_bug.cgi?id=2461107",
					Note: "Red Hat tracking bug",
				},
				{
					Type: "advisory",
					Url:  "https://ubuntu.com/security/CVE-2026-31532",
					Note: "Ubuntu CVE tracker entry",
				},
				{
					Type: "advisory",
					Url:  "https://osv.dev/vulnerability/CVE-2026-31532",
					Note: "OSV record with fix details",
				},
			},
			Boundaries: []string{
				"local_dos",
				"user_to_root",
				"container_to_host",
			},
			Mitigation: "* Block the CAN modules from loading on demand with lines in /etc/modprobe.d/disable-can.conf: `install can_raw /bin/false`, `install can /bin/false`\n* Block the virtual CAN drivers the same way: `install vcan /bin/false` and `install vxcan /bin/false` in modprobe.d\n* Disable unprivileged user namespaces with sysctl `user.max_user_namespaces=0` (or on Ubuntu `kernel.apparmor_restrict_unprivileged_userns=1`) so users cannot get CAP_NET_ADMIN to create a vcan\n* Run containers and pods with a RuntimeDefault seccomp profile that denies unshare/clone with CLONE_NEWUSER, and block socket(AF_CAN=29) where CAN is not needed",
		},
		Sync: vulnerability.Complete,
	},
	{
		Id:              "CVE-2026-31581",
		Source:          vulnerability.RedHat,
		Timestamp:       time.Date(2026, 9, 24, 5, 41, 34, 40000000, time.UTC),
		Published:       time.Date(2026, 4, 24, 0, 0, 0, 0, time.UTC),
		Status:          vulnerability.Analyzed,
		Description:     "In the Linux kernel, the following vulnerability has been resolved:\nALSA: 6fire: fix use-after-free on disconnect\nIn usb6fire_chip_abort(), the chip struct is allocated as the card's\nprivate data (via snd_card_new with sizeof(struct sfire_chip)).  When\nsnd_card_free_when_closed() is called and no file handles are open, the\ncard and embedded chip are freed synchronously.  The subsequent\nchip->card = NULL write then hits freed slab memory.\nCall trace:\nusb6fire_chip_abort sound/usb/6fire/chip.c:59 [inline]\nusb6fire_chip_disconnect+0x348/0x358 sound/usb/6fire/chip.c:182\nusb_unbind_interface+0x1a8/0x88c drivers/usb/core/driver.c:458\n...\nhub_event+0x1a04/0x4518 drivers/usb/core/hub.c:5953\nFix by moving the card lifecycle out of usb6fire_chip_abort() and into\nusb6fire_chip_disconnect().  The card pointer is saved in a local\nbefore any teardown, snd_card_disconnect() is called first to prevent\nnew opens, URBs are aborted while chip is still valid, and\nsnd_card_free_when_closed() is called last so chip is never accessed\nafter the card may be freed.\n\nA flaw was found in the Linux kernel's ALSA 6fire USB audio device driver. During the disconnection process of a 6fire USB audio device, a use-after-free vulnerability occurs. This happens when the system attempts to write to memory that has already been deallocated, which can lead to memory corruption and potentially cause system instability or a denial of service.",
		Statement:       "The 6fire USB audio disconnect path could touch freed memory; upstream hardens teardown ordering. Red Hat customers using this rare interface should update kernels. Optional mitigation is unloading the 6fire ALSA USB driver module where unused.",
		Score:           6.3,
		Severity:        vulnerability.Medium,
		Vector:          vulnerability.Local,
		Complexity:      vulnerability.High,
		Privileges:      vulnerability.Low,
		Interaction:     vulnerability.None,
		Scope:           vulnerability.Unchanged,
		Confidentiality: vulnerability.None,
		Integrity:       vulnerability.High,
		Availability:    vulnerability.High,
		Analysis: &vulnerability.Analysis{
			Updated:   time.Date(2026, 9, 24, 5, 41, 34, 40000000, time.UTC),
			Timestamp: time.Date(2026, 9, 24, 5, 21, 35, 234000000, time.UTC),
			Model:     "claude-opus-5-5",
			Summary:   "This flaw can't practically be reached on the server described. It triggers only when a TerraTec DMX 6Fire USB audio interface (USB ID 0ccd:0080), or a malicious USB device posing as one, is unplugged while the snd-usb-6fire driver is bound to it. Nobody can do that from a shell account, a container or a KVM guest (unless a matching USB device is passed through). No exploitation has been publicly reported. The bug is in sound/usb/6fire/chip.c. usb6fire_chip_abort() calls snd_card_free_when_closed(), and when no PCM or MIDI file handles are open, this frees the card and the sfire_chip struct stored inside it straight away. The function then writes chip->card = NULL into that freed slab memory. The upstream fix moves card teardown into usb6fire_chip_disconnect(). It saves the card pointer in a local variable, calls snd_card_disconnect() first, aborts the URBs while chip is still valid, and calls snd_card_free_when_closed() last. The bug was introduced by commit a0810c3d6dd2 ('ALSA: 6fire: Release resources at card release', November 2024). That commit was backported widely into stable kernels (5.4.287, 5.10.231 and later), so distribution kernels that took that backport are affected. Fixes landed in 6.12.83, 6.18.24, 6.19.14 and 7.0.1, with mainline commit 08e57f5e1a90. To trigger it, an attacker needs physical access, or control of USB passthrough or a USB/IP gadget, to attach and then detach a device matching the 6fire USB ID. A plain local user cannot trigger the disconnect path. The write is a single store of a NULL pointer into a freed object. That makes a crash or slab corruption realistic, while code execution would need precise heap grooming at hot-unplug time. The boundary crossed is physical access to a kernel crash (local DoS); the fix does not support 'low privilege local user' escalation. The snd-usb-6fire module is built on distribution kernels (Red Hat's own mitigation is to block it). It is autoloaded only by udev when matching USB hardware is enumerated, not by any unprivileged syscall, so the loaded-module check is a reliable signal. Red Hat rates the interface as rare and suggests blocking the module.",
			Impacted:  "* snd-usb-6fire ALSA driver for the TerraTec DMX 6Fire USB audio interface: use-after-free write when the device is unplugged\n* Hosts where USB devices can be physically attached, or where USB passthrough or USB/IP is set up, since a spoofed 6Fire device can trigger the disconnect bug",
			Score:     1.8,
			Universal: false,
			Processes: []string{},
			Modules: []string{
				"snd_usb_6fire",
			},
			Ports: []*vulnerability.Port{},
			Evidence: []*vulnerability.Evidence{
				{
					Type: "fix_commit",
					Url:  "https://git.kernel.org/stable/c/af75b486f7e883e3422ece23c8d727e6815144a0",
					Note: "Stable 6.12.83 fix: moves the snd_card_disconnect/snd_card_free_when_closed calls into usb6fire_chip_disconnect so chip is never touched after the card is freed",
				},
				{
					Type: "advisory",
					Url:  "https://ratatoskr.run/linux-cve-announce/2026/04/8048465",
					Note: "Kernel CVE team announcement: affected file sound/usb/6fire/chip.c; fixed in 6.12.83, 6.18.24, 6.19.14 and 7.0.1",
				},
				{
					Type: "other",
					Url:  "https://ratatoskr.run/cip-dev/2026/04/12477457/t",
					Note: "CIP weekly CVE report: introduced by a0810c3 ('ALSA: 6fire: Release resources at card release'); mainline fix 08e57f5e1a90",
				},
				{
					Type: "other",
					Url:  "https://opennet.ru/kernel/5.4.287.html",
					Note: "Shows the introducing commit a0810c3d6dd2 was backported to old stable branches (5.4.287), so older distribution kernels may carry the bug",
				},
				{
					Type: "advisory",
					Url:  "https://access.redhat.com/security/cve/cve-2026-31581",
					Note: "Red Hat: rare interface; mitigation is to prevent the snd_usb_6fire module from loading",
				},
				{
					Type: "advisory",
					Url:  "https://ubuntu.com/security/CVE-2026-31581",
					Note: "Ubuntu tracking entry, priority Medium",
				},
				{
					Type: "other",
					Url:  "https://cateee.net/lkddb/web-lkddb/SND_USB_6FIRE.html",
					Note: "CONFIG_SND_USB_6FIRE builds snd-usb-6fire and matches USB ID 0ccd:0080 only; loaded when the hardware is present",
				},
			},
			Boundaries: []string{
				"local_dos",
			},
			Mitigation: "* Block the driver: add 'install snd_usb_6fire /bin/false' to a file in /etc/modprobe.d/ and unload it with 'modprobe -r snd_usb_6fire' if it is loaded\n* Stop new USB devices from binding: set /sys/bus/usb/drivers_autoprobe to 0, or use USBGuard with a default-block policy\n* On KVM hosts, do not pass USB devices with ID 0ccd:0080 through to hosts or guests that don't need them",
		},
		Sync: vulnerability.Complete,
	},
	{
		Id:              "CVE-2026-43037",
		Source:          vulnerability.RedHat,
		Timestamp:       time.Date(2026, 9, 24, 5, 41, 33, 888000000, time.UTC),
		Published:       time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		Status:          vulnerability.Analyzed,
		Description:     "In the Linux kernel, the following vulnerability has been resolved:\nip6_tunnel: clear skb2->cb[] in ip4ip6_err()\nOskar Kjos reported the following problem.\nip4ip6_err() calls icmp_send() on a cloned skb whose cb[] was written\nby the IPv6 receive path as struct inet6_skb_parm. icmp_send() passes\nIPCB(skb2) to __ip_options_echo(), which interprets that cb[] region\nas struct inet_skb_parm (IPv4). The layouts differ: inet6_skb_parm.nhoff\nat offset 14 overlaps inet_skb_parm.opt.rr, producing a non-zero rr\nvalue. __ip_options_echo() then reads optlen from attacker-controlled\npacket data at sptr[rr+1] and copies that many bytes into dopt->__data,\na fixed 40-byte stack buffer (IP_OPTIONS_DATA_FIXED_SIZE).\nTo fix this we clear skb2->cb[], as suggested by Oskar Kjos.\nAlso add minimal IPv4 header validation (version == 4, ihl >= 5).\n\nA flaw was found in the Linux kernel's IPv6 tunnel implementation. A remote attacker could exploit this flaw by sending malicious ICMPv6 error messages to cause a stack-based buffer overflow in the kernel's IPv4-over-IPv6 tunnel error handling code. This could result in a kernel crash (denial of service) or potentially allow arbitrary code execution with kernel privileges.",
		Statement:       "This Critical flaw in the Linux kernel's IPv6 tunneling error handling can lead to a stack buffer overflow. An unauthenticated remote attacker could exploit this by sending specially crafted network packets, potentially resulting in a denial of service or information disclosure on affected Red Hat Enterprise Linux systems.",
		Score:           8.8,
		Severity:        vulnerability.Critical,
		Vector:          vulnerability.Network,
		Complexity:      vulnerability.Low,
		Privileges:      vulnerability.Low,
		Interaction:     vulnerability.None,
		Scope:           vulnerability.Unchanged,
		Confidentiality: vulnerability.High,
		Integrity:       vulnerability.High,
		Availability:    vulnerability.High,
		Analysis: &vulnerability.Analysis{
			Updated:   time.Date(2026, 9, 24, 5, 41, 33, 888000000, time.UTC),
			Timestamp: time.Date(2026, 9, 24, 5, 21, 56, 195000000, time.UTC),
			Model:     "claude-opus-5-5",
			Summary:   "On a default RHEL, AlmaLinux or Ubuntu install, an unprivileged local user can reach this flaw by creating a user and network namespace. No public exploit or in-the-wild exploitation has been reported since disclosure (fix posted to netdev on 26 March 2026, CVE published 1 May 2026; a June 2026 advisory from the University of Cologne found no public exploit code). The bug is in ip4ip6_err(), the ICMPv6 error handler for IPv4-in-IPv6 tunnels in the ip6_tunnel module (net/ipv6/ip6_tunnel.c). It has existed since 2.6.22. The handler clones the ICMPv6 error skb and calls icmp_send() without clearing skb->cb[]. The leftover IPv6 control block (inet6_skb_parm.nhoff) is then read as an IPv4 inet_skb_parm.opt.rr value. __ip_options_echo() takes an option length from attacker-supplied packet bytes and copies that many bytes into a fixed 40-byte stack buffer, so the overflow is an attacker-controlled kernel stack overwrite. The upstream fix (netdev commit 2edfa31769a4, backported to stable, e.g. 5.10.253, 5.15.203, 6.1.168) clears skb2->cb[] and checks that the inner IPv4 header has version 4 and ihl >= 5. **Trigger conditions:** an ICMPv6 error must arrive that quotes an IPv6 packet carrying IPPROTO_IPIP between addresses that match an UP ip6tnl tunnel in mode ipip6 or any. The inner IPv4 addresses must route back through a tunnel6 device. **Remote attackers:** exposure is limited to hosts that actually have such a tunnel configured, and the attacker must know or spoof the tunnel endpoint addresses. The result is a kernel crash (remote DoS). **Local attackers:** an unprivileged user can run unshare -Urn, then create and bring up an ip6tnl device. This makes the kernel request rtnl-link-ip6tnl, which autoloads ip6_tunnel with no capability check. The user can then inject the crafted ICMPv6 error inside their own netns with CAP_NET_RAW over loopback. So the code path is reachable on any default install that allows unprivileged user namespaces (RHEL 8/9/10 and derivatives). On Ubuntu 23.10 and later it is reachable only where the AppArmor unprivileged-userns restriction is not blocking it. **Full impact:** distribution kernels are built with CONFIG_STACK_PROTECTOR_STRONG, so an overwrite that reaches the canary panics the host. That makes a reliable host crash the practical outcome. Turning this into privilege escalation would require corrupting data in the icmp_send frame below the canary, or a canary leak. That is plausible in principle but not demonstrated, and remote code execution is highly unlikely. Red Hat rates the flaw Critical and has shipped fixes (e.g. RHSA-2026:27719, RHSA-2026:28742); Amazon Linux rates it Important, and Ubuntu and CIQ/Rocky 8/9/10 also track it. **Containers:** under default Docker/Podman/Kubernetes seccomp and capability policy, a container cannot create a nested user namespace or get CAP_NET_ADMIN to create or bring up the tunnel. Container-to-host impact (a host crash) therefore requires privileged containers, CAP_NET_ADMIN, or a runtime that permits unshare(CLONE_NEWUSER). **KVM guests** cannot reach the host path unless the host itself terminates ip4ip6 tunnels toward them. **Detection caveat:** ip6_tunnel is autoloadable, so a loaded-module check will not see it before an attack. A host showing ip6_tunnel loaded is using ip6tnl or ip6gre tunnels and is remotely exposed.",
			Impacted:  "* ip6_tunnel kernel module (ip6tnl devices): the ICMPv6 error handler for IPv4-in-IPv6 tunnels overflows a kernel stack buffer\n* Unprivileged user namespaces: let any local user autoload ip6_tunnel, create a tunnel and inject the malicious ICMPv6 error in a private netns\n* IPv4-in-IPv6 (4in6, DS-Lite style) tunnels in ipip6 or any mode: remote attackers who can spoof ICMPv6 errors toward the tunnel endpoints can crash the host\n* ip6_gre / ip6gretap tunnels: depend on ip6_tunnel, so hosts using them have the vulnerable module loaded\n* Privileged containers or pods with CAP_NET_ADMIN: can create or bring up ip6tnl in their netns and crash the shared host kernel",
			Score:     7,
			Universal: true,
			Processes: []string{},
			Modules: []string{
				"ip6_tunnel",
			},
			Ports: []*vulnerability.Port{},
			Evidence: []*vulnerability.Evidence{
				{
					Type: "fix_commit",
					Url:  "https://git.kernel.org/netdev/net/c/2edfa31769a4",
					Note: "Upstream fix by Eric Dumazet: clears skb2->cb[] in ip4ip6_err() and checks the inner IPv4 header (version 4, ihl >= 5); Fixes c4d3efafcc93 from 2.6.22",
				},
				{
					Type: "other",
					Url:  "https://ratatoskr.run/netdev/2026/03/11351618/t",
					Note: "netdev v2 patch thread: describes the IPv6/IPv4 cb[] layout confusion and the overflow of the 40-byte stack buffer in __ip_options_echo; applied to net on 2026-03-28",
				},
				{
					Type: "advisory",
					Url:  "https://ratatoskr.run/linux-cve-announce/2026/05/8048694",
					Note: "Kernel CVE announcement: introduced in 2.6.22, fixed in 5.10.253, 5.15.203, 6.1.168 and later stable releases",
				},
				{
					Type: "advisory",
					Url:  "https://access.redhat.com/security/cve/cve-2026-43037",
					Note: "Red Hat rates the flaw Critical and says it affects RHEL kernels",
				},
				{
					Type: "advisory",
					Url:  "https://access.redhat.com/errata/RHSA-2026:27719",
					Note: "RHEL kernel erratum fixing this CVE (BZ 2464351)",
				},
				{
					Type: "advisory",
					Url:  "https://access.redhat.com/errata/RHSA-2026:28742",
					Note: "Additional RHEL kernel erratum for this CVE",
				},
				{
					Type: "advisory",
					Url:  "https://ubuntu.com/security/CVE-2026-43037",
					Note: "Ubuntu tracking page for its kernels",
				},
				{
					Type: "advisory",
					Url:  "https://kb.ciq.com/article/security-advisories/cve-2026-43037-mitigation",
					Note: "CIQ/Rocky 8/9/10: exposure depends on ip6_tunnel being loaded; recommends the modprobe install /bin/false block",
				},
				{
					Type: "other",
					Url:  "https://itcc.uni-koeln.de/en/services/information-security/it-security/vulnerability-cve-2026-43037-on-linux",
					Note: "No public exploit code as of 12 June 2026; same module-block workaround",
				},
				{
					Type: "other",
					Url:  "https://explore.alas.aws.amazon.com/CVE-2026-43037.html",
					Note: "Amazon Linux rates it Important, CVSS 7.0",
				},
			},
			Boundaries: []string{
				"user_to_root",
				"local_dos",
				"remote_dos",
				"container_to_host",
			},
			Mitigation: "* Block the module: add `install ip6_tunnel /bin/false` to /etc/modprobe.d/ip6_tunnel.conf, then `rmmod ip6_tunnel` if it is loaded (this also disables ip6_gre and ip6tnl tunnels)\n* Stop unprivileged users from creating namespaces: set sysctl `user.max_user_namespaces=0` on RHEL/Alma; on Ubuntu use `kernel.apparmor_restrict_unprivileged_userns=1` or `kernel.unprivileged_userns_clone=0`\n* Keep CAP_NET_ADMIN and privileged mode away from untrusted containers, and keep the default seccomp profile that blocks unshare(CLONE_NEWUSER)\n* On hosts that need ip4ip6 tunnels: use ip6tables/nftables to drop ICMPv6 error messages (types 1-4) addressed to tunnel endpoints unless they come from trusted peers",
		},
		Sync: vulnerability.Complete,
	},
	{
		Id:              "CVE-2026-43501",
		Source:          vulnerability.RedHat,
		Timestamp:       time.Date(2026, 9, 24, 5, 21, 58, 324000000, time.UTC),
		Published:       time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC),
		Status:          vulnerability.Analyzed,
		Description:     "In the Linux kernel, the following vulnerability has been resolved:\nipv6: rpl: reserve mac_len headroom when recompressed SRH grows\nipv6_rpl_srh_rcv() decompresses an RFC 6554 Source Routing Header, swaps\nthe next segment into ipv6_hdr->daddr, recompresses, then pulls the old\nheader and pushes the new one plus the IPv6 header back.  The\nrecompressed header can be larger than the received one when the swap\nreduces the common-prefix length the segments share with daddr (CmprI=0,\nCmprE>0, seg[0][0] != daddr[0] gives the maximum +8 bytes).\npskb_expand_head() was gated on segments_left == 0, so on earlier\nsegments the push consumed unchecked headroom.  Once skb_push() leaves\nfewer than skb->mac_len bytes in front of data,\nskb_mac_header_rebuild()'s call to:\nskb_set_mac_header(skb, -skb->mac_len);\nwill store (data - head) - mac_len into the u16 mac_header field, which\nwraps to ~65530, and the following memmove() writes mac_len bytes ~64KiB\npast skb->head.\nA single AF_INET6/SOCK_RAW/IPV6_HDRINCL packet over lo with a two\nsegment type-3 SRH (CmprI=0, CmprE=15) reaches headroom 8 after one\npass; KASAN reports a 14-byte OOB write in ipv6_rthdr_rcv.\nFix this by expanding the head whenever the remaining room is less than\nthe push size plus mac_len, and request that much extra so the rebuilt\nMAC header fits afterwards.\n\nA flaw was found in the Linux kernel. A local attacker can exploit an out-of-bounds write vulnerability when the kernel recomputes an IPv6 Source Routing Header (SRH). This issue occurs because insufficient headroom is reserved during the recompression process, leading to memory corruption. Successful exploitation could result in a denial of service or potentially arbitrary code execution.",
		Statement:       "This is an Important flaw. A local attacker can exploit an out-of-bounds write in the Linux kernel's IPv6 Source Routing Header (SRH) recompression, which could lead to a denial of service or arbitrary code execution. This vulnerability impacts Red Hat Enterprise Linux 9 and 10, and Red Hat In-Vehicle OS 2, due to insufficient memory handling during SRH processing.",
		Score:           7.5,
		Severity:        vulnerability.High,
		Vector:          vulnerability.Adjacent,
		Complexity:      vulnerability.High,
		Privileges:      vulnerability.None,
		Interaction:     vulnerability.None,
		Scope:           vulnerability.Unchanged,
		Confidentiality: vulnerability.High,
		Integrity:       vulnerability.High,
		Availability:    vulnerability.High,
		Analysis: &vulnerability.Analysis{
			Updated:   time.Date(2026, 9, 24, 5, 21, 58, 324000000, time.UTC),
			Timestamp: time.Date(2026, 9, 24, 5, 21, 41, 960000000, time.UTC),
			Model:     "claude-opus-5-5",
			Summary:   "This flaw is practically reachable on a default RHEL 9/10, AlmaLinux 9/10 or Ubuntu server. Any unprivileged local user who can create a user and network namespace can trigger it. A full local-privilege-escalation exploit is public: NebuSec's \"Route of Root\" write-up and its CyberMeowfia repository, which won a kernelCTF bounty and are referenced by the Debian tracker since around the 2026 disclosure. The bug is in the RFC 6554 RPL Source Routing Header receive path, ipv6_rpl_srh_rcv() in net/ipv6/exthdrs.c. This code is part of core IPv6, is built into distribution kernels, is not a separate module, and has been present since v5.7. After swapping the next segment into daddr, the recompressed header can grow by up to 8 bytes. pskb_expand_head() was only called when segments_left==0, so skb_push() could consume the headroom unchecked. skb_mac_header_rebuild() then wraps the u16 mac_header, and memmove() writes mac_len (14) partially controlled bytes about 64KiB past skb->head. That is a deterministic slab out-of-bounds write that needs no race. The fix, commit 9e6bf146b559 in 7.1-rc2 and backported to stable, expands the head whenever the remaining headroom is less than the push size plus mac_len. The trigger needs RPL SRH processing, which is gated on net.ipv6.conf.all/<dev>.rpl_seg_enabled (default 0). In the host network namespace that gate means the path is not reachable by default. The CVSS \"adjacent/network\" framing only applies when an admin has set rpl_seg_enabled=1, which is rare outside 6LoWPAN/RPL deployments. The practical path is local: an unprivileged user runs unshare -Urn, becomes namespace root, sets rpl_seg_enabled=1 on lo in the new netns, and sends one AF_INET6/SOCK_RAW/IPV6_HDRINCL packet with a two-segment type-3 SRH over loopback. The resulting write can be groomed into arbitrary kernel read/write and root, crossing user to root. Full impact therefore needs unprivileged user namespaces (on by default on RHEL/AlmaLinux) plus CAP_NET_RAW and sysctl write access inside that namespace. On Ubuntu 24.04+ the AppArmor unprivileged-userns restriction adds a hurdle, but it has known bypasses. From a container, default Docker/Podman/Kubernetes seccomp policy blocks unshare(CLONE_NEWUSER) and /proc/sys is read-only. A default container therefore cannot enable the sysctl, and container escape needs a privileged container, a permitted userns, or a --sysctl setting. Affected: kernels 5.7 through pre-7.1. Red Hat lists RHEL 9 and 10 (RHEL 8 is not listed), and Ubuntu 22.04/24.04 and later kernels carry the code. No module is involved, so no module or port matching applies.",
			Impacted:  "* IPv6 extension-header receive path (RPL type-3 routing header): out-of-bounds write in ipv6_rpl_srh_rcv()\n* Unprivileged user + network namespaces (rootless tools, sandboxes): let any user enable rpl_seg_enabled in their own netns and reach the bug\n* Hosts with net.ipv6.conf.*.rpl_seg_enabled=1 (6LoWPAN/RPL routers): the bug becomes reachable by a single packet from the network\n* Privileged containers or pods allowed to set net.ipv6 sysctls or create user namespaces: can use the bug to escape to the host",
			Score:     9,
			Universal: true,
			Processes: []string{},
			Modules:   []string{},
			Ports:     []*vulnerability.Port{},
			Evidence: []*vulnerability.Evidence{
				{
					Type: "fix_commit",
					Url:  "https://git.kernel.org/linus/9e6bf146b55999a095bb14f73a843942456d1adc",
					Note: "Upstream fix (7.1-rc2): always expands the skb head when headroom is less than the push size plus mac_len in ipv6_rpl_srh_rcv()",
				},
				{
					Type: "fix_commit",
					Url:  "https://git.kernel.org/stable/c/4babc2d9fda2df43823b85d08a0180b68f1b0854",
					Note: "Stable backport of the fix",
				},
				{
					Type: "other",
					Url:  "https://ratatoskr.run/netdev/2026/04/3522616/t",
					Note: "netdev v3 patch thread; loopback raw-socket SRH trigger and KASAN 14-byte OOB write",
				},
				{
					Type: "exploitation",
					Url:  "https://nebusec.ai/research/cve-2026-43501-route-of-root/",
					Note: "Public LPE write-up (kernelCTF): affects 5.7-7.1; needs rpl_seg_enabled=1 or unprivileged user+net namespaces; partially controlled 14-byte OOB write turned into root",
				},
				{
					Type: "exploitation",
					Url:  "https://github.com/NebuSec/CyberMeowfia/tree/main/security-research/Linux-CVE-2026-43501",
					Note: "Public exploit source",
				},
				{
					Type: "advisory",
					Url:  "https://access.redhat.com/security/cve/cve-2026-43501",
					Note: "Red Hat: Important; affects RHEL 9 and 10",
				},
				{
					Type: "advisory",
					Url:  "https://ubuntu.com/security/CVE-2026-43501",
					Note: "Ubuntu tracker listing the fix and stable commits",
				},
				{
					Type: "advisory",
					Url:  "https://security-tracker.debian.org/tracker/CVE-2026-43501",
					Note: "Debian tracker referencing the fix and exploit",
				},
				{
					Type: "other",
					Url:  "https://www.spinics.net/lists/netdev/msg641215.html",
					Note: "Original RPL SRH patch: processing only happens if the rpl_seg_enabled sysctl is set",
				},
			},
			Boundaries: []string{
				"user_to_root",
				"container_to_host",
				"remote_to_local",
				"local_dos",
				"remote_dos",
			},
			Mitigation: "* Disable unprivileged user namespaces: sysctl user.max_user_namespaces=0 (RHEL/AlmaLinux), or on Ubuntu kernel.apparmor_restrict_unprivileged_userns=1 / kernel.unprivileged_userns_clone=0\n* Keep net.ipv6.conf.all.rpl_seg_enabled=0 and net.ipv6.conf.default.rpl_seg_enabled=0 on the host so the bug cannot be triggered remotely\n* Keep container seccomp profiles that deny unshare/clone with CLONE_NEWUSER, and do not grant containers CAP_SYS_ADMIN, --privileged, or --sysctl net.ipv6.conf.*.rpl_seg_enabled\n* If IPv6 is not needed, boot with ipv6.disable=1",
		},
		Sync: vulnerability.Complete,
	},
	{
		Id:              "CVE-2026-46054",
		Source:          vulnerability.RedHat,
		Timestamp:       time.Date(2026, 9, 24, 5, 21, 58, 119000000, time.UTC),
		Published:       time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		Status:          vulnerability.Analyzed,
		Description:     "In the Linux kernel, the following vulnerability has been resolved:\nselinux: fix overlayfs mmap() and mprotect() access checks\nThe existing SELinux security model for overlayfs is to allow access if\nthe current task is able to access the top level file (the \"user\" file)\nand the mounter's credentials are sufficient to access the lower\nlevel file (the \"backing\" file).  Unfortunately, the current code does\nnot properly enforce these access controls for both mmap() and mprotect()\noperations on overlayfs filesystems.\nThis patch makes use of the newly created security_mmap_backing_file()\nLSM hook to provide the missing backing file enforcement for mmap()\noperations, and leverages the backing file API and new LSM blob to\nprovide the necessary information to properly enforce the mprotect()\naccess controls.\n\nA flaw was found in the Linux kernel's SELinux security module when handling overlayfs. The existing security model for overlayfs does not properly enforce access controls for `mmap()` and `mprotect()` operations. This oversight could allow a local attacker to bypass intended security policies, potentially leading to unauthorized access to files within an overlayfs filesystem. The vulnerability is resolved by enhancing backing file enforcement for `mmap()` and leveraging new LSM (Linux Security Module) capabilities for `mprotect()` access controls.",
		Statement:       "",
		Score:           7,
		Severity:        vulnerability.High,
		Vector:          vulnerability.Local,
		Complexity:      vulnerability.High,
		Privileges:      vulnerability.Low,
		Interaction:     vulnerability.None,
		Scope:           vulnerability.Unchanged,
		Confidentiality: vulnerability.High,
		Integrity:       vulnerability.High,
		Availability:    vulnerability.High,
		Analysis: &vulnerability.Analysis{
			Updated:   time.Date(2026, 9, 24, 5, 21, 58, 119000000, time.UTC),
			Timestamp: time.Date(2026, 9, 24, 5, 21, 31, 878000000, time.UTC),
			Model:     "claude-opus-5-5",
			Summary:   "This flaw is reachable in principle on an SELinux-enforcing host that uses overlayfs, which covers any RHEL/AlmaLinux container host running podman, docker or Kubernetes. No exploitation has been reported, and the fix shows the impact is limited to bypassing parts of the SELinux policy, not memory corruption or privilege escalation. The bug is in security/selinux/hooks.c. SELinux's rule for overlayfs is that the calling task must be allowed to use the top-level (user) file and the mounter's credentials must be allowed to use the lower (backing) file. The mmap() path never ran the mounter-credential check against the backing file. The mprotect() path only had vma->vm_file, which is the backing file, so it checked the wrong object against the wrong credentials. Upstream fixed this with a three-patch series by Paul Moore and Amir Goldstein: new LSM hooks security_backing_file_alloc/free and security_mmap_backing_file, plus a per-backing_file LSM blob, and SELinux now checks both files. It landed in 7.1-rc1 (82544d36b172) and 7.0.4 (cd0e707a927a), and was backported to 6.12 LTS and to RHEL kernels (RHSA-2026:27811 and others; AlmaLinux ALSA-2026:27811). Ubuntu also tracks and patches it, but Ubuntu uses AppArmor rather than SELinux by default, so stock Ubuntu does not enforce SELinux and is effectively not exposed. To trigger it, an attacker needs a local process that can open a file on an overlayfs mount and then mmap() or mprotect() it, for example with PROT_EXEC or a shared writable mapping. What they get is an SELinux decision (execute, write-map or execmod on the lower file) that should have been denied because the mounter lacked the right. On typical container hosts the mounter is the container runtime or root, which usually has broad rights, so the missing check seldom changes the outcome. The exposure is larger when a confined, less-privileged domain mounts the overlay, for example an unprivileged user mounting overlayfs inside a user namespace. Even then the gain is getting around a mandatory-access-control restriction, not moving to root or escaping a container. That is why no privilege boundary is listed and why the CVSS C/I/A:H rating looks overstated. A later follow-up patch (Ondrej Mosnacek, July 2026) fixed a regression in which the new mounter check wrongly applied execmem, so patched kernels may need that as well. The overlay module is autoloaded when an overlay mount is made, including from an unprivileged user namespace.",
			Impacted:  "* SELinux enforcement on overlayfs mounts (RHEL/AlmaLinux default): mmap()/mprotect() never checked the mounter's rights to the lower file\n* Container root filesystems (podman, docker, CRI-O/Kubernetes on overlayfs): some container_t mapping and execute policy checks on image layers were incomplete\n* Rootless containers and unprivileged-user-namespace overlay mounts: the user acting as mounter can get past checks on the lower-layer file\n* FUSE passthrough and other backing-file users: they share the backing_file code the fix changes",
			Score:     3.5,
			Universal: true,
			Processes: []string{},
			Modules: []string{
				"overlay",
			},
			Ports: []*vulnerability.Port{},
			Evidence: []*vulnerability.Evidence{
				{
					Type: "other",
					Url:  "https://ratatoskr.run/linux-cve-announce/2026/05/17036835",
					Note: "Kernel CVE team announcement (mirror): fixed in 7.0.4 (cd0e707a927a) and 7.1-rc1 (82544d36b172); affected files security/selinux/hooks.c and objsec.h",
				},
				{
					Type: "fix_commit",
					Url:  "https://lore.kernel.org/all/20260403030848.731867-5-paul@paul-moore.com/",
					Note: "Upstream v4 patch 'selinux: fix overlayfs mmap() and mprotect() access checks', which uses the new security_mmap_backing_file hook and a backing_file LSM blob",
				},
				{
					Type: "other",
					Url:  "https://ratatoskr.run/linux-security-module/2026/04/346132/t",
					Note: "v4 cover letter: explains the missing mmap/mprotect checks on overlayfs and adds the backing_file alloc/free/mmap LSM hooks",
				},
				{
					Type: "other",
					Url:  "https://ratatoskr.run/linux-unionfs/2026/03/343618/t",
					Note: "Original series explaining why mprotect only sees the backing file in vma->vm_file, and the model of checking the user file plus the mounter's rights to the backing file",
				},
				{
					Type: "other",
					Url:  "https://ratatoskr.run/linux-security-module/2026/04/3508956/t",
					Note: "LSM pull request for 7.1 describing it as an important overlayfs bugfix for LSM mmap/mprotect controls",
				},
				{
					Type: "other",
					Url:  "https://ratatoskr.run/selinux/2026/06/17186016/t",
					Note: "Backport of the series to 6.12 LTS, tested with the selinux-testsuite overlay tests",
				},
				{
					Type: "other",
					Url:  "https://ratatoskr.run/selinux/2026/07/17210594/t",
					Note: "Follow-up patch fixing an incorrect execmem check the fix introduced in the mounter check",
				},
				{
					Type: "advisory",
					Url:  "https://access.redhat.com/errata/RHSA-2026:27811",
					Note: "Red Hat kernel erratum fixing the flaw (BZ 2482025), confirming RHEL kernels ship the affected code",
				},
				{
					Type: "advisory",
					Url:  "https://ubuntu.com/security/CVE-2026-46054",
					Note: "Ubuntu tracker; Ubuntu patches it but uses AppArmor by default, so SELinux is not enforced",
				},
				{
					Type: "advisory",
					Url:  "https://osv.dev/vulnerability/CVE-2026-46054",
					Note: "Lists distribution advisories: USN-8488-1, ALSA-2026:27811/27812/25191, openSUSE",
				},
			},
			Boundaries: []string{},
			Mitigation: "* Stop unprivileged users from mounting overlayfs by disabling user namespaces: sysctl user.max_user_namespaces=0 (this breaks rootless podman)\n* Where no containers or overlayfs are needed, add \"install overlay /bin/false\" to /etc/modprobe.d/ so the overlay module cannot load\n* Keep SELinux enforcing and review container_t, deny_execmem and execmod booleans; the missing checks apply only on overlayfs, not on regular filesystems",
		},
		Sync: vulnerability.Complete,
	},
	{
		Id:              "CVE-2026-42533",
		Source:          vulnerability.RedHat,
		Timestamp:       time.Date(2026, 9, 24, 8, 15, 29, 705000000, time.UTC),
		Published:       time.Date(2026, 7, 15, 14, 33, 45, 0, time.UTC),
		Status:          vulnerability.Analyzed,
		Description:     "A vulnerability exists in NGINX Plus and NGINX Open Source when a map directive uses regex matching and a string expression references the map's regex capture variables before referencing the map output variable. Alternatively, the same result could be achieved by using a non-cacheable variable in a string expression under certain conditions. An unauthenticated attacker along with conditions beyond their control can exploit this vulnerability by sending crafted HTTP requests. This may cause a heap buffer overflow in the NGINX worker process leading to a restart. Additionally, attackers can execute code on systems with Address Space Layout Randomization (ASLR) disabled or when the attacker can bypass ASLR.\nImpact:\nThis vulnerability may allow remote attackers to cause a denial-of-service (DoS) on the NGINX system or to possibly trigger a code execution. There is no control plane exposure; this is a data plane issue only.\n Note: Software versions which have reached End of Technical Support (EoTS) are not evaluated.\n\nA flaw was found in NGINX. An unauthenticated attacker can exploit this vulnerability by sending specially crafted HTTP requests when the `map` directive uses regular expression (regex) matching and references regex capture variables before referencing the map output variable. This can lead to a heap buffer overflow, which may allow for arbitrary code execution on systems where Address Space Layout Randomization (ASLR) is disabled or bypassed. Additionally, this flaw can cause a denial-of-service (DoS) due to the NGINX worker process restarting.",
		Statement:       "This vulnerability in NGINX allows a remote, unauthenticated attacker to trigger a heap buffer overflow, leading to a denial of service (Dos). This occurs when the `map` directive uses regex matching and references regex capture variables before the map output variable. While arbitrary code execution is a theoretical risk, it is significantly reduced on Red Hat systems where Address Space Layout Randomization (ASLR) is enabled by default.",
		Score:           7,
		Severity:        vulnerability.High,
		Vector:          vulnerability.Network,
		Complexity:      vulnerability.High,
		Privileges:      vulnerability.None,
		Interaction:     vulnerability.None,
		Scope:           vulnerability.Unchanged,
		Confidentiality: vulnerability.Low,
		Integrity:       vulnerability.Low,
		Availability:    vulnerability.High,
		Analysis: &vulnerability.Analysis{
			Updated:   time.Date(2026, 9, 24, 8, 15, 29, 705000000, time.UTC),
			Timestamp: time.Date(2026, 9, 24, 8, 15, 29, 705000000, time.UTC),
			Model:     "claude-opus-5-5",
			Summary:   "The flaw is reachable only when nginx runs a specific configuration. A string expression has to reference a regex capture ($1..$9, or a named capture) before a regex-backed map variable that overwrites the same capture state, or it has to include a non-cacheable ('volatile') variable whose length changes between evaluation passes. A default package install does not hit this, but regex maps next to regex locations, server_name, rewrite and if captures are common in production reverse-proxy and ingress configs. The bug is in the nginx script engine (http and stream modules; every version from 0.9.6 through 1.30.3 and 1.31.2, fixed in 1.30.4 and 1.31.3, plus NGINX Plus R33 to R36 P6 and 37.0.0.1 to 37.0.2.1). The engine builds strings in two passes: a length pass sizes the buffer, then a copy pass writes into it. The upstream fix, commit b767540 ('Script: buffer overrun protection', with follow-up patches for the proxy, fastcgi, scgi, uwsgi, grpc, index, try_files and access_log direct evaluators), adds a bound on writes past the allocation made in the length pass, and makes the final length count only the bytes actually written. Its commit message shows the trigger configurations: a named-capture map used after its capture, and a volatile map. The primitive is therefore a heap buffer overflow in the worker process where the attacker controls both the contents and the size of the write. After $1 has been sized, a lazily evaluated map regex over request data (URI, headers, or TLS SNI in stream with ssl_preread) replaces the capture array, so the copy pass writes the longer, attacker-supplied capture into the short buffer. The reporters' write-ups describe a write of about 12,000 bytes under ASan. When the new capture is shorter, uninitialised heap memory is left inside the reported length and can come back in responses or headers, which leaks heap pointers. The reporters describe chaining that leak with the overflow to get past ASLR, so ASLR makes exploitation less reliable but does not rule out code execution, despite Red Hat lowering the score to 7.0 on that basis. An unauthenticated remote attacker only needs to send HTTP requests (or TLS ClientHellos to a stream listener) to a vulnerable location or server block. Crossing the boundary gives worker crashes (repeated restarts, DoS), heap disclosure, and possibly code execution as the nginx worker user. Distribution nginx builds are all in the affected range: RHEL/AlmaLinux 8 module streams (1.14 to 1.24), RHEL 9 (1.20 plus the 1.22, 1.24 and 1.26 streams), RHEL 10 (1.26), and supported Ubuntu releases. Ubuntu shipped a fix in USN-8563-1, pulled it twice over ABI and regression problems with third-party modules, and issued the final fix in USN-8563-5. Anything that embeds the nginx core is affected, including the ingress-nginx controller, OpenResty and Tengine, which all run as 'nginx' processes. No distribution kernel or container-runtime policy is involved. F5 says named captures are only a partial workaround, because a named-capture clobbering variant also exists.",
			Impacted:  "* nginx http module: worker heap overflow when a regex map variable follows capture references ($1, named captures) in set, return, proxy_pass, add_header, rewrite and similar directives\n* nginx stream module: the same overflow in return and similar directives when a regex map runs over attacker-controlled data such as ssl_preread SNI\n* Kubernetes ingress-nginx and other nginx-based ingress/reverse-proxy stacks (OpenResty, Tengine) that generate regex maps and regex locations\n* nginx maps marked volatile (non-cacheable variables) used inside string expressions\n* access_log with variables: the same unbounded copy, fixed by a follow-up commit\n* NGINX Plus R33 to R36 P6 and 37.0.0.1 to 37.0.2.1",
			Score:     8.2,
			Universal: false,
			Processes: []string{
				"nginx",
			},
			Modules: []string{},
			Ports: []*vulnerability.Port{
				{Name: "http", Protocol: "tcp", Port: 80},
				{Name: "https", Protocol: "tcp", Port: 443},
				{Name: "http-alt", Protocol: "tcp", Port: 8080},
				{Name: "https-alt", Protocol: "tcp", Port: 8443},
				{Name: "http3", Protocol: "udp", Port: 443},
			},
			Evidence: []*vulnerability.Evidence{
				{
					Type: "fix_commit",
					Url:  "https://github.com/nginx/nginx/commit/b767540492e8c79a58bc26034d3bab2f708b7bd1",
					Note: "Upstream fix 'Script: buffer overrun protection'. It bounds copy-pass writes to the length-pass allocation in ngx_http_complex_value, ngx_http_script_run and the regex/complex value codes, and does the same in stream. The commit message shows the trigger configs (a named-capture map used after its capture, and a volatile map).",
				},
				{
					Type: "other",
					Url:  "https://github.com/nginx/nginx/pull/1561",
					Note: "The 1.31.3 security PR. It also adds the overrun protection to the direct evaluators (proxy, fastcgi, scgi, uwsgi, grpc, index, try_files) and to access_log.",
				},
				{
					Type: "advisory",
					Url:  "https://nginx.org/en/CHANGES-1.30",
					Note: "Upstream changelog: heap buffer overflow with a regex map after a capture, or with non-cacheable variables, fixed in 1.30.4.",
				},
				{
					Type: "advisory",
					Url:  "https://github.com/advisories/GHSA-fxfg-v4rj-hp95",
					Note: "GitHub advisory entry for this CVE (CWE-122): heap overflow in the worker, reachable unauthenticated with a vulnerable configuration.",
				},
				{
					Type: "advisory",
					Url:  "https://access.redhat.com/security/cve/cve-2026-42533",
					Note: "Red Hat rates it 7.0 (C:L/I:L/A:H), treating ASLR as the reason code execution is unlikely. Affects RHEL nginx packages.",
				},
				{
					Type: "advisory",
					Url:  "https://ubuntu.com/security/CVE-2026-42533",
					Note: "Ubuntu tracker: supported releases affected. The original patches changed ABI, and a similar fix was also applied to access log.",
				},
				{
					Type: "advisory",
					Url:  "https://ubuntu.com/security/notices/USN-8563-5",
					Note: "The final Ubuntu fix, after the fix was reverted in USN-8563-2 and USN-8563-4 over ABI and regression problems.",
				},
				{
					Type: "other",
					Url:  "https://winfunc.com/hacktivity/CVE-2026-42533",
					Note: "Reporter write-up (stream variant): a return \"$1$m\" expression with an SNI-driven regex map gave an attacker-controlled 12,000-byte write past a one-byte allocation under ASan.",
				},
				{
					Type: "other",
					Url:  "https://cyberstan.co.uk/nginx-rce/",
					Note: "Co-reporter's root-cause write-up: PCRE capture state is not saved and restored across the two script passes. The attacker controls overflow content and length through the URI, headers or body. The same bug gives a heap info leak, and many directive call sites are affected in both http and stream.",
				},
				{
					Type: "other",
					Url:  "https://1w4y.io/blog/nginx-over-under-copy/",
					Note: "Root-cause analysis: a longer clobbered capture overruns the buffer, while a shorter one leaves uninitialised bytes inside the recorded length. The patch bounds copies and fixes the final length.",
				},
			},
			Boundaries: []string{
				"remote_to_local",
				"remote_dos",
				"info_disclosure",
			},
			Mitigation: "* In every block that uses a regex map, do not reference $1..$9 (or a named capture) earlier in the same string expression than the map variable. Copy captures into ordinary variables first (for example `set $p $1;`) before evaluating the map variable\n* Give map regexes named captures (`~(?<mapcap>...)`) that are distinct from, and never reused in, location, server_name, rewrite and if regexes. F5 notes this is only a partial workaround\n* Remove the `volatile;` parameter from map blocks whose output is used inside string expressions",
		},
		Sync: vulnerability.Complete,
	},
	{
		Id:              "CVE-2026-74480",
		Source:          vulnerability.RedHat,
		Timestamp:       time.Date(2026, 9, 17, 23, 51, 56, 861000000, time.UTC),
		Published:       time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
		Status:          vulnerability.Analyzed,
		Description:     "In the Linux kernel, the following vulnerability has been resolved:\nnet: bridge: stop fast-leave after deleting a port group\nbr_multicast_leave_group() iterates mp->ports with pp = &p->next in\nits fast-leave path. After br_multicast_del_pg() removes p,\ncontinuing the loop advances pp through the deleted entry.\nIf multicast-to-unicast was enabled, the bridge can hold multiple port\ngroups for the same port and group with different source MAC\naddresses. Once multicast-to-unicast is disabled,\nbr_port_group_equal() matches those entries by port only. A fast leave\ncan then delete one entry and continue from its stale next pointer,\nleaving mp->ports pointing at a deleted port group.\nFast leave only needs to remove one matching port group. Break after\nbr_multicast_del_pg() so the loop stops before dereferencing the\nremoved entry.\n\nA flaw was found in the Linux kernel's network bridge module. When handling multicast fast-leave, a vulnerability exists where the system may attempt to access a port group after it has been deleted. This can occur if multicast-to-unicast was previously enabled and then disabled, leading to a stale pointer. This memory corruption vulnerability could lead to system instability or a denial of service.",
		Statement:       "If the multicast-to-unicast feature has not been explicitly enabled on any bridge port (it is disabled by default), the vulnerable code path cannot be reached. Systems that do not use Linux bridge interfaces - for example, systems that are not acting as hypervisors, container hosts, or network gateways - are not affected by this flaw.",
		Score:           7.8,
		Severity:        vulnerability.High,
		Vector:          vulnerability.Local,
		Complexity:      vulnerability.Low,
		Privileges:      vulnerability.Low,
		Interaction:     vulnerability.None,
		Scope:           vulnerability.Unchanged,
		Confidentiality: vulnerability.High,
		Integrity:       vulnerability.High,
		Availability:    vulnerability.High,
		Analysis: &vulnerability.Analysis{
			Updated:   time.Date(2026, 9, 17, 23, 51, 56, 862000000, time.UTC),
			Timestamp: time.Date(2026, 9, 15, 6, 30, 23, 994000000, time.UTC),
			Model:     "claude-opus-5",
			Summary:   "Reachable only on hosts with a non-default Linux bridge port configuration, and no public exploitation has been reported as of this writing. The flaw is in the bridge multicast snooping code: in br_multicast_leave_group(), the fast-leave path walks mp->ports with pp = &p->next and keeps iterating after br_multicast_del_pg() has unlinked and kfree_rcu'd the matching entry; the upstream fix (Fixes: 6db6f0eae605 \"bridge: multicast to unicast\", one-line \"break\" after br_multicast_del_pg()) stops the loop. Two entries for the same port can only exist if mcast_to_unicast was enabled on that port (br_port_group_equal() then also matches the source MAC), and both entries only match again once mcast_to_unicast is turned off; deleting the first then re-splicing through the stale &p->next leaves mp->ports pointing at a freed port group, i.e. a genuine use-after-free on the MDB list. Trigger conditions: the bridge module must be loaded and in use (docker0, virbr0, KVM/OVS-less VM bridges, container CNI bridges), multicast snooping must be on (default on), the port must have fastleave/mcast_fast_leave enabled (default OFF per ip-link(8)/bridge(8)), and mcast_to_unicast must have been enabled on that port and later disabled (also default OFF). Given that, the attacker only needs L2 access to a bridged segment - a KVM guest, a container on a bridged CNI, or any host on the same broadcast domain - and sends IGMP/MLD joins from several source MACs followed by an IGMPv2 leave; no local shell account or capability on the host is required. Additional conditions gating full impact: turning the dangling mp->ports entry into anything beyond a crash requires heap grooming of the net_bridge_port_group slab across an RCU grace period, which is far harder than the reliable list-corruption panic; realistically this is a kernel-memory-corruption DoS with a theoretical path to host kernel code execution from a guest/container. All currently supported distributions ship the vulnerable code (RHEL 9 and RHEL 10 kernels were fixed in RHSA-2026:62568/62609 and RHSA-2026:66355; Ubuntu and Debian kernels are likewise affected); CONFIG_BRIDGE is built as the autoloadable 'bridge' module in all of them, and an unprivileged user in a user namespace can autoload it, but the flaw still needs the two non-default per-port flags, so universal is false and the module list is the practical host match. The Red Hat vendor statement is accurate on the mcast_to_unicast gate; the CVSS 7.8 local/AV:L vector understates the attack vector (L2 remote/guest) but overstates the ease, since fastleave plus a mcast_to_unicast on->off transition is an unusual combination on server bridges (it is mostly seen on wireless AP/bridge appliances).",
			Impacted:  "* bridge kernel module - use-after-free in br_multicast_leave_group() fast-leave path corrupts the MDB port-group list\n* KVM/libvirt host bridges (virbr0, br0 with tap interfaces) - a guest on the bridge can send the IGMP sequence that triggers the stale pointer\n* Docker/Podman/Kubernetes bridge networking (docker0, cni0, CNI bridge plugin) - containers sit on bridge ports and can reach the same code path\n* Bridge port flag mcast_to_unicast - must have been enabled and then disabled for duplicate port groups to become matchable\n* Bridge port flag fastleave / mcast_fast_leave - gates entry into the vulnerable loop; default off\n* Bridge multicast snooping (multicast_snooping=1, default) - required for IGMP/MLD leave processing to reach br_multicast_leave_group()\n* IGMPv2/MLD leave handling in net/bridge/br_multicast.c - the exact code changed by the fix\n* Network gateway/appliance hosts bridging untrusted L2 segments - exposed to any station on the segment",
			Score:     5.2,
			Universal: false,
			Processes: []string{},
			Modules: []string{
				"bridge",
			},
			Ports: []*vulnerability.Port{},
			Evidence: []*vulnerability.Evidence{
				{
					Type: "other",
					Url:  "http://www.mail-archive.com/bridge@lists.linux.dev/msg01607.html",
					Note: "Upstream patch '[PATCH net 1/1] net: bridge: stop fast-leave after deleting a port group' - single-line break after br_multicast_del_pg() in net/bridge/br_multicast.c, Fixes: 6db6f0eae605 ('bridge: multicast to unicast'), Cc stable.",
				},
				{
					Type: "advisory",
					Url:  "https://access.redhat.com/security/cve/cve-2026-74480",
					Note: "Red Hat CVE page: bridge module flaw, states the mcast_to_unicast enable-then-disable precondition and characterises impact as memory corruption / denial of service.",
				},
				{
					Type: "advisory",
					Url:  "https://access.redhat.com/errata/RHSA-2026:62609",
					Note: "RHSA fixing CVE-2026-74480 in RHEL kernel (BZ 2517046), confirming shipped distribution kernels contain the vulnerable code.",
				},
				{
					Type: "advisory",
					Url:  "https://access.redhat.com/errata/RHSA-2026:62568",
					Note: "Additional RHEL kernel erratum listing the same bridge fast-leave fix, showing multiple supported RHEL streams are affected.",
				},
				{
					Type: "advisory",
					Url:  "https://ubuntu.com/security/CVE-2026-74480",
					Note: "Ubuntu security tracker entry for the same bridge fast-leave use-after-free, confirming Ubuntu kernels ship the affected code.",
				},
				{
					Type: "other",
					Url:  "https://man7.org/linux/man-pages/man8/ip-link.8.html",
					Note: "ip-link(8) bridge_slave documentation: mcast_to_unicast and fastleave are both off by default, establishing that the trigger requires non-default port configuration.",
				},
			},
			Boundaries: []string{
				"remote_to_local",
				"guest_to_host",
				"container_to_host",
				"remote_dos",
				"local_dos",
			},
			Mitigation: "* Disable the vulnerable path on every bridge port: `ip link set dev <port> type bridge_slave fastleave off` (mcast_fast_leave=0); without this flag br_multicast_leave_group() never enters the buggy loop.\n* Never enable `mcast_to_unicast on` on bridge ports (`ip link set dev <port> type bridge_slave mcast_to_unicast off`); duplicate per-MAC port groups cannot be created without it, and if it was previously enabled, flush the MDB (`bridge mdb flush dev br0`) after disabling it.\n* Where IGMP snooping is not needed, turn it off: `ip link set dev br0 type bridge mcast_snooping 0`, which stops leave messages from reaching the multicast port-group code.\n* On hosts that use no bridges at all, block the module: `install bridge /bin/false` in /etc/modprobe.d/ to prevent unprivileged autoload.",
		},
		Sync: vulnerability.Complete,
	},
}

func advisoryResourceState(reference string, resourceId bson.ObjectID) (
	state string, exclusions []string) {

	state = advisory.Affected
	exclusions = []string{}

	switch reference {
	case "ALSA-2026:25191":
		exclusions = []string{"CVE-2026-31467", "CVE-2026-31581"}
	case "ALSA-2026:67314":
		web := false
		for _, info := range advisoryInstancesInfo {
			if info.Id == resourceId && info.Name == "web-app" {
				web = true
				break
			}
		}
		if !web {
			state = advisory.Unreachable
			exclusions = []string{"CVE-2026-42533"}
		}
	case "ALSA-2026:66355":
		state = advisory.Unreachable
		exclusions = []string{"CVE-2026-74480"}
	}

	return
}

func GetAdvisoryDetail(advId bson.ObjectID) *aggregate.AdvisoryDetail {
	var adv *advisory.Advisory
	for _, a := range Advisories {
		if a.Id == advId {
			adv = a
			break
		}
	}
	if adv == nil {
		return nil
	}

	vulns := []*vulnerability.Vulnerability{}
	for _, cveId := range adv.Vulnerabilities {
		for _, vuln := range advisoryVulnerabilities {
			if vuln.Id == cveId {
				vulns = append(vulns, vuln)
				break
			}
		}
	}

	counts := &aggregate.AdvisoryCounts{}
	exclusions := map[string]int64{}

	instancesInfo := []*aggregate.AdvisoryInstanceInfo{}
	for _, info := range advisoryInstancesInfo {
		state, excl := advisoryResourceState(adv.Reference, info.Id)

		infoCopy := *info
		infoCopy.State = state
		instancesInfo = append(instancesInfo, &infoCopy)

		counts.Active += 1
		if state == advisory.Unreachable {
			counts.UnreachableInstances += 1
		} else {
			counts.Instances += 1
		}
		for _, cveId := range excl {
			exclusions[cveId] += 1
		}
	}

	nodesInfo := []*aggregate.AdvisoryNodeInfo{}
	for _, info := range advisoryNodesInfo {
		state, excl := advisoryResourceState(adv.Reference, info.Id)

		infoCopy := *info
		infoCopy.State = state
		nodesInfo = append(nodesInfo, &infoCopy)

		counts.Active += 1
		if state == advisory.Unreachable {
			counts.UnreachableNodes += 1
		} else {
			counts.Nodes += 1
		}
		for _, cveId := range excl {
			exclusions[cveId] += 1
		}
	}

	return &aggregate.AdvisoryDetail{
		Id:              adv.Id,
		Vulnerabilities: vulns,
		InstancesInfo:   instancesInfo,
		NodesInfo:       nodesInfo,
		Counts:          counts,
		Exclusions:      exclusions,
		Page:            0,
		PageCount:       20,
		Count:           counts.Active,
	}
}

func GetResourceAdvisories(
	resourceId bson.ObjectID) []*aggregate.ResourceAdvisory {

	advisories := []*aggregate.ResourceAdvisory{}

	kind := ""
	for _, info := range advisoryInstancesInfo {
		if info.Id == resourceId {
			kind = advisory.Instance
			break
		}
	}
	for _, info := range advisoryNodesInfo {
		if info.Id == resourceId {
			kind = advisory.Node
			break
		}
	}
	if kind == "" {
		return advisories
	}

	for _, adv := range Advisories {
		detail := GetAdvisoryDetail(adv.Id)
		state, exclusions := advisoryResourceState(adv.Reference, resourceId)

		advisories = append(advisories, &aggregate.ResourceAdvisory{
			Advisory:          *adv,
			VulnerabilityDocs: detail.Vulnerabilities,
			Resource: &advisory.Resource{
				Organization: adv.Organization,
				Reference:    adv.Reference,
				Resource:     resourceId,
				Kind:         kind,
				State:        state,
				Dismissed:    false,
				Exclusions:   exclusions,
				Timestamp:    adv.Updated,
			},
		})
	}

	return advisories
}
