package settings

var Hypervisor *hypervisor

type hypervisor struct {
	Id           string `bson:"_id"`
	SystemdPath  string `bson:"systemd_path" default:"/etc/systemd/system"`
	LibPath      string `bson:"systemd_path" default:"/var/lib/pritunl-cloud"`
	BridgeName   string `bson:"bridge_name" default:"pritunlbr0"`
	NormalMtu    int    `bson:"normal_mtu" default:"1500"`
	JumboMtu     int    `bson:"jumbo_mtu" default:"9000"`
	StartTimeout int    `bson:"start_timeout" default:"30"`
	StopTimeout  int    `bson:"stop_timeout" default:"120"`
	RefreshRate  int    `bson:"refresh_rate" default:"90"`
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
