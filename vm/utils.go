package vm

import (
	"bytes"
	"crypto/md5"
	"encoding/base32"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

func GetMacAddr(id bson.ObjectId, secondId bson.ObjectId) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hash.Write([]byte(secondId.Hex()))
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

func GetMacAddrVirt(id bson.ObjectId, secondId bson.ObjectId) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hash.Write([]byte(secondId.Hex()))
	macHash := fmt.Sprintf("%x", hash.Sum(nil))
	macHash = macHash[:10]
	macBuf := bytes.Buffer{}

	for i, run := range macHash {
		macBuf.WriteRune(run)
		if i%2 == 1 && i != len(macHash)-1 {
			macBuf.WriteRune(':')
		}
	}

	return "02:" + macBuf.String()
}

func GetIface(id bson.ObjectId, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("p%s%d", strings.ToLower(hashSum), n)
}

func GetIfaceVirt(id bson.ObjectId, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("v%s%d", strings.ToLower(hashSum), n)
}

func GetIfaceInternal(id bson.ObjectId, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("i%s%d", strings.ToLower(hashSum), n)
}

func GetIfaceVlan(id bson.ObjectId, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("x%s%d", strings.ToLower(hashSum), n)
}

func GetNamespace(id bson.ObjectId, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("n%s%d", strings.ToLower(hashSum), n)
}
