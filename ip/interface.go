package ip

type Interface struct {
	Name    string `bson:"name" json:"name"`
	Address string `bson:"address" json:"address"`
	Master  string `bson:"master" json:"master"`
}
