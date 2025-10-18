package settings

var Hypervisor *hypervisor

type hypervisor struct {
	Id                     string `bson:"_id"`
	SystemdPath            string `bson:"systemd_path" default:"/etc/systemd/system"`
	LibPath                string `bson:"lib_path" default:"/var/lib/pritunl-cloud"`
	RunPath                string `bson:"run_path" default:"/var/run/pritunl-cloud"`
	AgentHostPath          string `bson:"agent_host_path" default:"/usr/bin/pritunl-cloud-agent"`
	AgentBsdHostPath       string `bson:"agent_bsd_host_path" default:"/usr/bin/pritunl-cloud-agent-bsd"`
	AgentGuestPath         string `bson:"agent_guest_path" default:"/usr/bin/pci"`
	InitGuestPath          string `bson:"init_guest_path" default:"/etc/pritunl-cloud-init"`
	HugepagesPath          string `bson:"hugepages_path" default:"/dev/hugepages/pritunl"`
	LockCloudPass          bool   `bson:"lock_cloud_pass"`
	DesktopEnv             string `bson:"desktop_env" default:"gnome"`
	OvmfCodePath           string `bson:"ovmf_code_path"`
	OvmfVarsPath           string `bson:"ovmf_vars_path"`
	OvmfSecureCodePath     string `bson:"ovmf_secure_code_path"`
	OvmfSecureVarsPath     string `bson:"ovmf_secure_vars_path"`
	NbdPath                string `bson:"nbd_path" default:"/dev/nbd6"`
	DiskAio                string `bson:"disk_aio"`
	NoSandbox              bool   `bson:"no_sandbox"`
	GlHostMem              int    `bson:"gl_host_mem" default:"2048"`
	BridgeIfaceName        string `bson:"bridge_iface_name" default:"br0"`
	ImdsIfaceName          string `bson:"imds_iface_name" default:"imds0"`
	NormalMtu              int    `bson:"normal_mtu" default:"1500"`
	JumboMtu               int    `bson:"jumbo_mtu" default:"9000"`
	DiskQueuesMin          int    `bson:"disk_queues_min" default:"1"`
	DiskQueuesMax          int    `bson:"disk_queues_max" default:"4"`
	NetworkQueuesMin       int    `bson:"network_queues_min" default:"1"`
	NetworkQueuesMax       int    `bson:"network_queues_max" default:"8"`
	CloudInitNetVer        int    `bson:"cloud_init_net_ver" default:"1"`
	HostNetwork            string `bson:"host_network" default:"198.18.84.0/22"`
	HostNetworkName        string `bson:"host_network_name" default:"pritunlhost0"`
	VirtRng                bool   `bson:"virt_rng"`
	VlanRanges             string `bson:"vlan_ranges" default:"1001-3999"`
	VxlanId                int    `bson:"vxlan_id" default:"9417"`
	VxlanDestPort          int    `bson:"vxlan_dest_port" default:"4789"`
	IpTimeout              int    `bson:"ip_timeout" default:"30"`
	IpTimeout6             int    `bson:"ip_timeout6" default:"15"`
	ActionRate             int    `bson:"action_rate" default:"3"`
	NodePortNetwork        string `bson:"node_port_network" default:"198.19.96.0/23"`
	NodePortRanges         string `bson:"node_port_ranges" default:"30000-32767"`
	NodePortNetworkName    string `bson:"node_port_network_name" default:"pritunlport0"`
	AddressRefreshTtl      int    `bson:"address_refresh_ttl" default:"1800"`
	StartTimeout           int    `bson:"start_timeout" default:"45"`
	StopTimeout            int    `bson:"stop_timeout" default:"180"`
	RefreshRate            int    `bson:"refresh_rate" default:"90"`
	SplashTime             int    `bson:"splash_time" default:"60"`
	DhcpRenewTtl           int    `bson:"dhcp_renew_ttl" default:"60"`
	NoIpv6PingInit         bool   `bson:"no_ipv6_ping_init"`
	Ipv6PingHost           string `bson:"ipv6_ping_host" default:"2001:4860:4860::8888"`
	ImdsAddress            string `bson:"imds_address" default:"169.254.169.254/32"`
	ImdsPort               int    `bson:"imds_port" default:"80"`
	ImdsSyncLogTimeout     int    `bson:"imds_sync_log_timeout" default:"20"`
	ImdsSyncRestartTimeout int    `bson:"imds_sync_log_timeout" default:"30"`
	InfoTtl                int    `bson:"info_ttl" default:"10"`
	NoGuiFullscreen        bool   `bson:"no_gui_fullscreen"`
	UsbHsPorts             int    `bson:"usb_hs_ports" default:"4"`
	UsbSsPorts             int    `bson:"usb_ss_ports" default:"4"`
	NoVirtioHid            bool   `bson:"no_virtio_hid"`
	JournalDisplayLimit    int64  `bson:"journal_display_limit" default:"3000"`
	DhcpLifetime           int    `bson:"dhcp_lifetime" default:"3600"`
	NdpRaInterval          int    `bson:"ndp_ra_interval" default:"6"`
	DnsServerPrimary       string `bson:"dns_server_primary" default:"8.8.8.8"`
	DnsServerSecondary     string `bson:"dns_server_secondary" default:"8.8.4.4"`
	DnsServerPrimary6      string `bson:"dns_server_primary6" default:"2001:4860:4860::8888"`
	DnsServerSecondary6    string `bson:"dns_server_secondary6" default:"2001:4860:4860::8844"`
	NodePortMaxAttempts    int    `bson:"node_port_max_attempts" default:"10000"`
	MaxDeploymentFailures  int    `bson:"max_deployment_failures" default:"3"`
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
