package settings

var System *system

type system struct {
	Id                   string `bson:"_id"`
	Name                 string `bson:"name"`
	DatabaseVersion      int    `bson:"database_version"`
	Demo                 bool   `bson:"demo"`
	License              string `bson:"license"`
	AdminCookieAuthKey   []byte `bson:"admin_cookie_auth_key"`
	AdminCookieCryptoKey []byte `bson:"admin_cookie_crypto_key"`
	UserCookieAuthKey    []byte `bson:"user_cookie_auth_key"`
	UserCookieCryptoKey  []byte `bson:"user_cookie_crypto_key"`
	NodeTimestampTtl     int    `bson:"node_timestamp_ttl" default:"15"`
	InstanceTimestampTtl int    `bson:"instance_timestamp_ttl" default:"10"`
	DomainLockTtl        int    `bson:"domain_lock_ttl" default:"30"`
	DomainDeleteTtl      int    `bson:"domain_delete_ttl" default:"200"`
	DomainRefreshTtl     int    `bson:"domain_refresh_ttl" default:"200"`
	AcmeKeyAlgorithm     string `bson:"acme_key_algorithm" default:"rsa"`
	DiskBackupWindow     int    `bson:"disk_backup_window" default:"6"`
	DiskBackupTime       int    `bson:"disk_backup_time" default:"10"`
	PlannerBatchSize     int    `bson:"planner_batch_size" default:"10"`
	NoMigrateRefresh     bool   `bson:"no_migrate_refresh"`
	OracleApiRetryRate   int    `bson:"oracle_api_retry_rate" default:"1"`
	OracleApiRetryCount  int    `bson:"oracle_api_retry_count" default:"120"`
	TwilioAccount        string `bson:"twilio_account"`
	TwilioSecret         string `bson:"twilio_secret"`
	TwilioNumber         string `bson:"twilio_number"`
}

func newSystem() interface{} {
	return &system{
		Id: "system",
	}
}

func updateSystem(data interface{}) {
	System = data.(*system)
}

func init() {
	register("system", newSystem, updateSystem)
}
