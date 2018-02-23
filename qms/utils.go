package qms

import (
	"fmt"
	"github.com/pritunl/pritunl-cloud/settings"
	"gopkg.in/mgo.v2/bson"
	"path"
)

func GetSockPath(virtId bson.ObjectId) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.sock", virtId.Hex()))
}
