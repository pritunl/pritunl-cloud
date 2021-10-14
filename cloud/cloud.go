package cloud

type Subnet struct {
	Id      string `bson:"id" json:"id"`
	VpcId   string `bson:"vpc_id" json:"vpc_id"`
	Name    string `bson:"name" json:"name"`
	Network string `bson:"network" json:"network"`
}

type Vpc struct {
	Id      string    `bson:"id" json:"id"`
	Name    string    `bson:"name" json:"name"`
	Network string    `bson:"network" json:"network"`
	Subnets []*Subnet `bson:"subnets" json:"subnets"`
}
