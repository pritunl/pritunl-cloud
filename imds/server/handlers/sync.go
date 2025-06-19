package handlers

import (
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/imds/server/config"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/server/state"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	lastSecurity     = time.Now().Add(-7 * time.Minute)
	lastSecurityLock sync.Mutex
)

type syncRespData struct {
	Spec string `json:"spec"`
	Hash uint32 `json:"hash"`
}

func syncPut(c *gin.Context) {
	data := &types.State{}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	if !state.Global.State.Final() {
		state.Global.State.Status = data.Status
	}
	state.Global.State.Timestamp = time.Now()
	state.Global.State.Memory = data.Memory
	state.Global.State.HugePages = data.HugePages
	state.Global.State.Load1 = data.Load1
	state.Global.State.Load5 = data.Load5
	state.Global.State.Load15 = data.Load15

	if data.Security != nil {
		if len(data.Security.Updates) > 50 {
			data.Security.Updates = data.Security.Updates[:50]
		}
		state.Global.State.Security = data.Security
	}

	if data.Output != nil {
		for _, entry := range data.Output {
			state.Global.AppendOutput(entry)
		}
	}

	if data.Hash != config.Config.Hash {
		c.JSON(200, &syncRespData{
			Spec: config.Config.SpecData,
			Hash: config.Config.Hash,
		})
	} else {
		c.JSON(200, &syncRespData{
			Hash: config.Config.Hash,
		})
	}
}

func hostSyncPut(c *gin.Context) {
	data := &types.Config{}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	if data.Hash != 0 {
		config.Config = data
	}

	ste := state.Global.State.Copy()
	ste.Output = state.Global.GetOutput()

	lastSecurityLock.Lock()
	if time.Since(lastSecurity) > 10*time.Minute {
		lastSecurity = time.Now()
	} else {
		ste.Security = nil
	}
	lastSecurityLock.Unlock()

	c.JSON(200, ste)
}

func hostSyncGet(c *gin.Context) {
	ste := state.Global.State.Copy()
	ste.Output = state.Global.GetOutput()

	lastSecurityLock.Lock()
	if time.Since(lastSecurity) > 10*time.Minute {
		lastSecurity = time.Now()
	} else {
		ste.Security = nil
	}
	lastSecurityLock.Unlock()

	c.JSON(200, ste)
}
