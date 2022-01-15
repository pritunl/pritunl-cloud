package settings

var Hypervisor *hypervisor

type hypervisor struct {
	Id                 string `bson:"_id"`
	SystemdPath        string `bson:"systemd_path" default:"/etc/systemd/system"`
	LibPath            string `bson:"systemd_path" default:"/var/lib/pritunl-cloud"`
	RunPath            string `bson:"run_path" default:"/var/run/pritunl-cloud"`
	HugepagesPath      string `bson:"hugepages_path" default:"/dev/hugepages/pritunl"`
	DesktopEnv         string `bson:"desktop_env" default:"gnome"`
	OvmfCodePath       string `bson:"ovmf_vars_path"`
	OvmfVarsPath       string `bson:"ovmf_vars_path"`
	OvmfSecureCodePath string `bson:"ovmf_secure_vars_path"`
	OvmfSecureVarsPath string `bson:"ovmf_secure_vars_path"`
	DiskAio            string `bson:"disk_aio"`
	NoSandbox          bool   `bson:"no_sandbox"`
	NormalMtu          int    `bson:"normal_mtu" default:"1500"`
	JumboMtu           int    `bson:"jumbo_mtu" default:"9000"`
	DiskQueuesMin      int    `bson:"disk_queues_min" default:"1"`
	DiskQueuesMax      int    `bson:"disk_queues_max" default:"4"`
	NetworkQueuesMin   int    `bson:"network_queues_min" default:"1"`
	NetworkQueuesMax   int    `bson:"network_queues_max" default:"8"`
	VxlanId            int    `bson:"vxlan_id" default:"9417"`
	VxlanDestPort      int    `bson:"vxlan_dest_port" default:"4789"`
	IpTimeout          int    `bson:"ip_timeout" default:"30"`
	HostNetworkName    string `bson:"host_network_name" default:"pritunlhost0"`
	StartTimeout       int    `bson:"start_timeout" default:"45"`
	StopTimeout        int    `bson:"stop_timeout" default:"180"`
	RefreshRate        int    `bson:"refresh_rate" default:"90"`
	SplashTime         int    `bson:"splash_time" default:"60"`
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
