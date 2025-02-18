package vm

import (
	"bytes"
	"crypto/md5"
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
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

func GetMacAddrExternal6(id primitive.ObjectID,
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

	return "08:" + macBuf.String()
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

func GetMacAddrHost(id primitive.ObjectID,
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

	return "06:" + macBuf.String()
}

func GetMacAddrNodePort(id primitive.ObjectID,
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

	return "0a:" + macBuf.String()
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

func GetIfaceHost(id primitive.ObjectID, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("h%s%d", strings.ToLower(hashSum), n)
}

func GetIfaceOracle(id primitive.ObjectID, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("o%s%d", strings.ToLower(hashSum), n)
}

func GetIfaceOracleVirt(id primitive.ObjectID, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("t%s%d", strings.ToLower(hashSum), n)
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

func GetHostVxlanIface(parentIface string) string {
	hash := md5.New()
	hash.Write([]byte(parentIface))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("k%s0", strings.ToLower(hashSum))
}

func GetHostBridgeIface(parentIface string) string {
	hash := md5.New()
	hash.Write([]byte(parentIface))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:12]
	return fmt.Sprintf("b%s0", strings.ToLower(hashSum))
}
