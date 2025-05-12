package settings

var Acme *acme

type acme struct {
	Id                string `bson:"_id"`
	Url               string `bson:"url" default:"https://acme-v01.api.letsencrypt.org"`
	DnsMaxConcurrent  int    `bson:"dns_max_concurrent" default:"10"`
	DnsRetryRate      int    `bson:"dns_retry_rate" default:"3"`
	DnsTimeout        int    `bson:"dns_timeout" default:"45"`
	DnsDelay          int    `bson:"dns_delay" default:"15"`
	DnsAwsTtl         int    `bson:"dns_aws_ttl" default:"3"`
	DnsCloudflareTtl  int    `bson:"dns_cloudflare_ttl" default:"60"`
	DnsOracleCloudTtl int    `bson:"dns_oracle_cloud_ttl" default:"3"`
}

func newAcme() interface{} {
	return &acme{
		Id: "acme",
	}
}

func updateAcme(data interface{}) {
	Acme = data.(*acme)
}

func init() {
	register("acme", newAcme, updateAcme)
}
