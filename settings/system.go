package settings

const (
	SetOnInsert = "set_on_insert"
	Merge       = "merge"
	Overwrite   = "overwrite"
)

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
	AcmeKeyAlgorithm     string `bson:"acme_key_algorithm" default:"rsa"`
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
