package data

import (
	"crypto/md5"
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

func GetVgName(id primitive.ObjectID, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:9]
	return fmt.Sprintf("cvg_%s%d", strings.ToLower(hashSum), n)
}

func GetLvName(id primitive.ObjectID, n int) string {
	hash := md5.New()
	hash.Write([]byte(id.Hex()))
	hashSum := base32.StdEncoding.EncodeToString(hash.Sum(nil))[:9]
	return fmt.Sprintf("clv_%s%d", strings.ToLower(hashSum), n)
}
