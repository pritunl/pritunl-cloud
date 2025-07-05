package settings

var Router *router

type router struct {
	Id                   string `bson:"_id"`
	ReadTimeout          int    `bson:"read_timeout" default:"300"`
	ReadHeaderTimeout    int    `bson:"read_header_timeout" default:"60"`
	WriteTimeout         int    `bson:"write_timeout" default:"300"`
	IdleTimeout          int    `bson:"idle_timeout" default:"60"`
	DialTimeout          int    `bson:"dial_timeout" default:"60"`
	DialKeepAlive        int    `bson:"dial_keep_alive" default:"60"`
	MaxIdleConns         int    `bson:"max_idle_conns" default:"1000"`
	MaxIdleConnsPerHost  int    `bson:"max_idle_conns_per_host" default:"100"`
	IdleConnTimeout      int    `bson:"idle_conn_timeout" default:"90"`
	HandshakeTimeout     int    `bson:"handshake_timeout" default:"10"`
	ContinueTimeout      int    `bson:"continue_timeout" default:"10"`
	MaxHeaderBytes       int    `bson:"max_header_bytes" default:"4194304"`
	ForceRedirectSystemd bool   `bson:"force_redirect_systemd"`
	SkipVerify           bool   `bson:"skip_verify"`
	ProxyResolverRefresh int    `bson:"proxy_resolver_refresh" default:"30"`
	ProxyResolverTtl     int    `bson:"proxy_resolver_ttl" default:"30"`
}

func newRouter() interface{} {
	return &router{
		Id: "router",
	}
}

func updateRouter(data interface{}) {
	Router = data.(*router)
}

func init() {
	register("router", newRouter, updateRouter)
}
