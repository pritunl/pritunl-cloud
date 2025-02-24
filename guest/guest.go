package guest

import (
	"time"

	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	socketsLock = utils.NewMultiTimeoutLock(1 * time.Minute)
)

type Command struct {
	Execute   string                 `json:"execute"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type Response struct {
	Return  map[string]interface{} `json:"return"`
	Error   map[string]interface{} `json:"error,omitempty"`
	RawData []byte                 `json:"-"`
}
