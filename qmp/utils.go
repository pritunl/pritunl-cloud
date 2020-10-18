package qmp

import (
	"fmt"
	"path"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/settings"
)

func GetQmpSockPath(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.qmp.sock", virtId.Hex()))
}
