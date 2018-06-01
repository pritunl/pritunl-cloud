package link

import (
	"gopkg.in/mgo.v2/bson"
	"sync"
)

var (
	Hashes     = map[bson.ObjectId]string{}
	HashesLock = sync.Mutex{}
	LinkStatus = map[bson.ObjectId]Status{}
	LinkStatusLock = sync.Mutex{}
)

type Status map[string]map[string]string

type State struct {
	Id          string        `json:"id"`
	VpcId       bson.ObjectId `json:"-"`
	Ipv6        bool          `json:"ipv6"`
	Type        string        `json:"type"`
	Secret      string        `json:"-"`
	Hash        string        `json:"hash"`
	Links       []*Link       `json:"links"`
	PublicAddr  string        `json:"-"`
	PublicAddr6 string        `json:"-"`
}

type Link struct {
	PreSharedKey string   `json:"pre_shared_key"`
	Right        string   `json:"right"`
	LeftSubnets  []string `json:"left_subnets"`
	RightSubnets []string `json:"right_subnets"`
}
