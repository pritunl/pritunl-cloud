package handlers

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/dhcpc"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/server/state"
	"github.com/pritunl/pritunl-cloud/utils"
)

func dhcpPut(c *gin.Context) {
	data := &dhcpc.Lease{}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	state.Global.SetDhcp(data.Iface, data.Iface6,
		data.Address, data.Address6, data.Gateway, data.Timestamp)

	c.JSON(200, map[string]string{})
}
