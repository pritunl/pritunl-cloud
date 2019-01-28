package link

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"sync"
)

var (
	Hashes         = map[primitive.ObjectID]string{}
	HashesLock     = sync.Mutex{}
	LinkStatus     = map[primitive.ObjectID]Status{}
	LinkStatusLock = sync.Mutex{}
)

type Status map[string]map[string]string

type State struct {
	Id          string             `json:"id"`
	VpcId       primitive.ObjectID `json:"-"`
	Ipv6        bool               `json:"ipv6"`
	Type        string             `json:"type"`
	Secret      string             `json:"-"`
	Hash        string             `json:"hash"`
	Links       []*Link            `json:"links"`
	PublicAddr  string             `json:"-"`
	PublicAddr6 string             `json:"-"`
}

type Link struct {
	PreSharedKey string   `json:"pre_shared_key"`
	Right        string   `json:"right"`
	LeftSubnets  []string `json:"left_subnets"`
	RightSubnets []string `json:"right_subnets"`
}
