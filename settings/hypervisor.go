package settings

var Hypervisor *hypervisor

type hypervisor struct {
	Id                  string `bson:"_id"`
	SystemdPath         string `bson:"systemd_path" default:"/etc/systemd/system"`
	LibPath             string `bson:"lib_path" default:"/var/lib/pritunl-cloud"`
	RunPath             string `bson:"run_path" default:"/var/run/pritunl-cloud"`
	AgentHostPath       string `bson:"cli_host_path" default:"/usr/bin/pritunl-cloud-agent"`
	CliGuestPath        string `bson:"cli_guest_path" default:"/usr/bin/pci"`
	InitGuestPath       string `bson:"init_guest_path" default:"/etc/pritunl-cloud-init"`
	HugepagesPath       string `bson:"hugepages_path" default:"/dev/hugepages/pritunl"`
	DesktopEnv          string `bson:"desktop_env" default:"gnome"`
	OvmfCodePath        string `bson:"ovmf_code_path"`
	OvmfVarsPath        string `bson:"ovmf_vars_path"`
	OvmfSecureCodePath  string `bson:"ovmf_secure_code_path"`
	OvmfSecureVarsPath  string `bson:"ovmf_secure_vars_path"`
	DiskAio             string `bson:"disk_aio"`
	NoSandbox           bool   `bson:"no_sandbox"`
	NormalMtu           int    `bson:"normal_mtu" default:"1500"`
	JumboMtu            int    `bson:"jumbo_mtu" default:"9000"`
	DiskQueuesMin       int    `bson:"disk_queues_min" default:"1"`
	DiskQueuesMax       int    `bson:"disk_queues_max" default:"4"`
	NetworkQueuesMin    int    `bson:"network_queues_min" default:"1"`
	NetworkQueuesMax    int    `bson:"network_queues_max" default:"8"`
	CloudInitNetVer     int    `bson:"cloud_init_net_ver" default:"1"`
	VirtRng             bool   `bson:"virt_rng"`
	VxlanId             int    `bson:"vxlan_id" default:"9417"`
	VxlanDestPort       int    `bson:"vxlan_dest_port" default:"4789"`
	IpTimeout           int    `bson:"ip_timeout" default:"30"`
	IpTimeout6          int    `bson:"ip_timeout6" default:"15"`
	HostNetworkName     string `bson:"host_network_name" default:"pritunlhost0"`
	StartTimeout        int    `bson:"start_timeout" default:"45"`
	StopTimeout         int    `bson:"stop_timeout" default:"180"`
	RefreshRate         int    `bson:"refresh_rate" default:"90"`
	SplashTime          int    `bson:"splash_time" default:"60"`
	NoIpv6PingInit      bool   `bson:"no_ipv6_ping_init"`
	ImdsAddress         string `bson:"imds_address" default:"169.254.169.254/16"`
	ImdsPort            int    `bson:"imds_port" default:"80"`
	ImdsLogDisplayLimit int64  `bson:"imds_log_display_limit" default:"5000"`
	KmsgLogDisplayLimit int64  `bson:"kmsg_log_display_limit" default:"5000"`
	DnsServerPrimary    string `bson:"dns_server_primary" default:"8.8.8.8"`
	DnsServerSecondary  string `bson:"dns_server_secondary" default:"8.8.4.4"`
	DnsServerPrimary6   string `bson:"dns_server_primary6" default:"2001:4860:4860::8888"`
	DnsServerSecondary6 string `bson:"dns_server_secondary6" default:"2001:4860:4860::8844"`
}

func newHypervisor() interface{} {
	return &hypervisor{
		Id: "hypervisor",
	}
}

func updateHypervisor(data interface{}) {
	Hypervisor = data.(*hypervisor)
}

func init() {
	register("hypervisor", newHypervisor, updateHypervisor)
}
