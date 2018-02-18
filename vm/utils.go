package vm

import (
	"bytes"
	"crypto/md5"
	"encoding/base32"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"path"
	"strings"
)

func GetVmPath(id bson.ObjectId) string {
	return path.Join(Root, id.Hex())
}

func GetDiskPath(id bson.ObjectId, num int) string {
	return path.Join(GetVmPath(id), fmt.Sprintf("disk%d.img", num))
}

func GetMacAddr(id bson.ObjectId) string {
	hash := md5.New()
	hash.Write([]byte(id))
	macHash := fmt.Sprintf("%x", hash.Sum(nil))
	macHash = macHash[:10]
	macBuf := bytes.Buffer{}

	for i, run := range macHash {
		macBuf.WriteRune(run)
		if i%2 == 1 && i != 9 {
			macBuf.WriteRune(':')
		}
	}

	return "00:" + macBuf.String()
}

func GetIface(id bson.ObjectId, n int) string {
	hash := md5.New()
	hash.Write([]byte(id))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("p%s%d", strings.ToLower(hashSum), n)
}
