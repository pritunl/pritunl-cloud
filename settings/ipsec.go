package settings

var Ipsec *ipsec

type ipsec struct {
	Id                         string `bson:"_id"`
	LinkTimeout                int    `bson:"link_timeout" default:"15"`
	DisconnectedTimeout        int    `bson:"disconnected_timeout" default:"60"`
	DisableDisconnectedRestart bool   `bson:"disable_disconnected_restart"`
	StateCacheTtl              int    `bson:"state_cache_ttl" default:"25"`
	SkipVerify                 bool   `bson:"skip_verify"`
}

func newIpsec() interface{} {
	return &ipsec{
		Id: "ipsec",
	}
}

func updateIpsec(data interface{}) {
	Ipsec = data.(*ipsec)
}

func init() {
	register("ipsec", newIpsec, updateIpsec)
}
