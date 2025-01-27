package spec

type Domain struct {
	Records []*Record `bson:"record" json:"record"`
}

type Record struct {
	Name   string      `bson:"name" json:"name"`
	Values []*Refrence `bson:"values" json:"values"`
}

type DomainYaml struct {
	Name    string             `yaml:"name"`
	Kind    string             `yaml:"kind"`
	Records []DomainYamlRecord `yaml:"records"`
}

type DomainYamlRecord struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}
