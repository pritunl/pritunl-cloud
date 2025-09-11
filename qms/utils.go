package qms

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/utils"
)

// TODO Backward compatibility
func GetSockPath(virtId bson.ObjectID) (pth string, err error) {
	sockPath := paths.GetSockPath(virtId)
	sockPathOld := paths.GetSockPathOld(virtId)

	exists, err := utils.Exists(sockPath)
	if err != nil {
		return
	}

	if exists {
		pth = sockPath
	} else {
		pth = sockPathOld
	}

	return
}
