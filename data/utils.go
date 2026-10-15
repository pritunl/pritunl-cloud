package data

import (
	"crypto/md5"
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/disk"
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
