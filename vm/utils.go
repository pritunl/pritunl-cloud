package vm

import (
	"bytes"
	"crypto/md5"
	"encoding/base32"
	"fmt"
	"github.com/pritunl/pritunl-cloud/node"
	"gopkg.in/mgo.v2/bson"
	"path"
	"strings"
)

func GetVmPath(instId bson.ObjectId) string {
	return path.Join(node.Self.GetVirtPath(),
		"instances", instId.Hex())
}

func GetDisksPath() string {
	return path.Join(node.Self.GetVirtPath(), "disks")
}

func GetDiskPath(diskId bson.ObjectId) string {
	return path.Join(GetDisksPath(),
		fmt.Sprintf("%s.qcow2", diskId.Hex()))
}

func GetMacAddr(id bson.ObjectId) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	macHash := fmt.Sprintf("%x", hash.Sum(nil))
	macHash = macHash[:10]
	macBuf := bytes.Buffer{}

	for i, run := range macHash {
		macBuf.WriteRune(run)
		if i%2 == 1 && i != len(macHash)-1 {
			macBuf.WriteRune(':')
		}
	}

	return "00:" + macBuf.String()
}

func GetIface(id bson.ObjectId, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("p%s%d", strings.ToLower(hashSum), n)
}
