package vm

import (
	"bytes"
	"crypto/md5"
	"encoding/base32"
	"fmt"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"strings"
)

func GetMacAddr(id primitive.ObjectID, secondId primitive.ObjectID) string {
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

func GetMacAddrExternal(id primitive.ObjectID,
	secondId primitive.ObjectID) string {

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

func GetMacAddrInternal(id primitive.ObjectID,
	secondId primitive.ObjectID) string {

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

	return "04:" + macBuf.String()
}

func GetIface(id primitive.ObjectID, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("p%s%d", strings.ToLower(hashSum), n)
}

func GetIfaceVirt(id primitive.ObjectID, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("v%s%d", strings.ToLower(hashSum), n)
}

func GetIfaceExternal(id primitive.ObjectID, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("e%s%d", strings.ToLower(hashSum), n)
}

func GetIfaceInternal(id primitive.ObjectID, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("i%s%d", strings.ToLower(hashSum), n)
}

func GetIfaceVlan(id primitive.ObjectID, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("x%s%d", strings.ToLower(hashSum), n)
}

func GetNamespace(id primitive.ObjectID, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("n%s%d", strings.ToLower(hashSum), n)
}

func GetLinkIfaceExternal(id primitive.ObjectID, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("z%s%d", strings.ToLower(hashSum), n)
}

func GetLinkIfaceInternal(id primitive.ObjectID, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("y%s%d", strings.ToLower(hashSum), n)
}

func GetLinkIfaceVirt(id primitive.ObjectID, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("w%s%d", strings.ToLower(hashSum), n)
}

func GetLinkNamespace(id primitive.ObjectID, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("x%s%d", strings.ToLower(hashSum), n)
}
