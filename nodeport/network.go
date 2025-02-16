package nodeport

import (
	"net"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/requires"
	"github.com/pritunl/pritunl-cloud/settings"
)

var (
	network *net.IPNet
)

func init() {
	module := requires.New("nodeport")
	module.After("settings")

	module.Handler = func() (err error) {
		_, network, err = net.ParseCIDR(settings.Hypervisor.NodePortNetwork)
		if err != nil {
			err = &errortypes.ParseError{
				errors.Wrap(err,
					"nodeport: Failed to parse node port network"),
			}
			return
		}

		return
	}
}
