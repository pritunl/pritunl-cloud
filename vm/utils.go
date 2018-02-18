package vm

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"path"
)

func GetVmPath(id bson.ObjectId) string {
	return path.Join(Root, id.Hex())
}

func GetDiskPath(id bson.ObjectId, num int) string {
	return path.Join(GetVmPath(id), fmt.Sprintf("disk%d.img", num))
}

func GetMacAddr(id bson.ObjectId) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	macHash := fmt.Sprintf("%x", hash.Sum(nil))
	macHash = macHash[:12]
	macBuf := bytes.Buffer{}

	for i, run := range macHash {
		macBuf.WriteRune(run)
		if i%2 == 1 && i != 11 {
			macBuf.WriteRune(':')
		}
	}

	return macBuf.String()
}
