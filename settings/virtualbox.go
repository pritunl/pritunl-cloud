package settings

var VirtualBox *virtualBox

type virtualBox struct {
	Id              string `bson:"_id"`
	ManagePath      string `bson:"manage_path" default:"VBoxManage"`
	PowerOffTimeout int    `bson:"power_off_timeout" default:"160"`
}

func newVirtualBox() interface{} {
	return &virtualBox{
		Id: "virtual_box",
	}
}

func updateVirtualBox(data interface{}) {
	VirtualBox = data.(*virtualBox)
}

func init() {
	register("virtual_box", newVirtualBox, updateVirtualBox)
}
