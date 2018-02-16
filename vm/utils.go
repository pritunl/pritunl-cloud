package vm

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"path"
)

func GetVmPath(id bson.ObjectId) string {
	return path.Join(Root, id.Hex())
}

func GetDiskPath(id bson.ObjectId, num int) string {
	return path.Join(GetVmPath(id), fmt.Sprintf("disk%d.vmdk", num))
}

func GetMacAddr(id bson.ObjectId) string {
	return "08" + id.Hex()[14:]
}
