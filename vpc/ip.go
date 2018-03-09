package vpc

import "gopkg.in/mgo.v2/bson"

type VpcIp struct {
	Id       bson.ObjectId `bson:"_id,omitempty"`
	Vpc      bson.ObjectId `bson:"vpc"`
	Ip       int64         `bson:"ip"`
	Instance bson.ObjectId `bson:"instance"`
}
