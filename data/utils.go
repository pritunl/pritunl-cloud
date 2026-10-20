package data

import (
	"crypto/md5"
	"encoding/base32"
	"fmt"
	"strings"
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/sirupsen/logrus"
)

func GetVgName(id bson.ObjectID, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:9]
	return fmt.Sprintf("%s%s%d", disk.VgNamePrefix,
		strings.ToLower(hashSum), n)
}

func IsManagedVgName(name string) bool {
	return strings.HasPrefix(name, disk.VgNamePrefix)
}

func GetLvName(id bson.ObjectID, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:9]
	return fmt.Sprintf("%s%s%d", disk.LvNamePrefix,
		strings.ToLower(hashSum), n)
}

func ShutdownWait() {
	start := time.Now()
	logged := false

	for FrozenInstances() > 0 {
		if time.Since(start) > 90*time.Second {
			logrus.WithFields(logrus.Fields{
				"count": FrozenInstances(),
			}).Error("cmd.node: Guest filesystems still frozen, exiting")
			return
		}

		if !logged {
			logged = true
			logrus.WithFields(logrus.Fields{
				"count": FrozenInstances(),
			}).Info("cmd.node: Waiting for guest filesystems to thaw")
		}

		time.Sleep(100 * time.Millisecond)
	}
}
