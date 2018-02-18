package settings

var Qemu *qemu

type qemu struct {
	Id           string `bson:"_id"`
	SystemdPath  string `bson:"systemd_path" default:"/etc/systemd/system"`
	LibPath      string `bson:"systemd_path" default:"/var/lib/pritunl-cloud"`
	StartTimeout int    `bson:"start_timeout" default:"30"`
	StopTimeout  int    `bson:"stop_timeout" default:"120"`
}

func newQemu() interface{} {
	return &qemu{
		Id: "qemu",
	}
}

func updateQemu(data interface{}) {
	Qemu = data.(*qemu)
}

func init() {
	register("qemu", newQemu, updateQemu)
}
