package telemetry

const (
	Host      = "host"
	Instance  = "instance"
	Namespace = "namespace"

	RedHat  = "rhel"
	FreeBsd = "freebsd"

	moderate  = "moderate"
	important = "important"
	critical  = "critical"

	Low      = 1
	Medium   = 2
	High     = 3
	Critical = 4
)

var (
	Mode                 = Host
	NetworkInternalIface = ""
	NetworkExternalIface = ""
)
