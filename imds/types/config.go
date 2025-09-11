package types

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/utils"
)

// Cannot contain maps for encode order
type Config struct {
	Spec           bson.ObjectID  `json:"spec"`
	SpecData       string         `json:"spec_data" gob:"-"`
	ImdsHostSecret string         `json:"-"`
	ClientIps      []string       `json:"client_ips"`
	Node           *Node          `json:"node"`
	Instance       *Instance      `json:"instance"`
	Vpc            *Vpc           `json:"vpc"`
	Subnet         *Subnet        `json:"subnet"`
	Certificates   []*Certificate `json:"certificates"`
	Secrets        []*Secret      `json:"secrets"`
	Pods           []*Pod         `json:"pods"`
	Hash           uint32         `json:"hash"`
}

func (c *Config) ComputeHash() (err error) {
	c.Hash = 0

	confHash, err := utils.CrcHash(c)
	if err != nil {
		return
	}

	c.Hash = confHash
	return
}
