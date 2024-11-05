package handlers

import (
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/imds/server/config"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/server/state"
	"github.com/pritunl/pritunl-cloud/utils"
)

type syncData struct {
	Memory    float64 `json:"memory"`
	HugePages float64 `json:"hugepages"`
	Load1     float64 `json:"load1"`
	Load5     float64 `json:"load5"`
	Load15    float64 `json:"load15"`
}

type syncRespData struct {
	Hash uint32 `json:"hash"`
}

func syncPut(c *gin.Context) {
	data := &syncData{}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	state.State.Timestamp = time.Now()
	state.State.Memory = data.Memory
	state.State.HugePages = data.HugePages
	state.State.Load1 = data.Load1
	state.State.Load5 = data.Load5
	state.State.Load15 = data.Load15

	err = state.State.Save()
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, &syncRespData{
		Hash: config.Config.Hash,
	})
}
