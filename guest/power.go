package guest

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
)

const (
	shutdownTimeout = 5 * time.Second
)

func Shutdown(vmId bson.ObjectID) (err error) {
	_, err = executeOpt(vmId, &Command{
		Execute: "guest-shutdown",
		Arguments: map[string]interface{}{
			"mode": "powerdown",
		},
	}, shutdownTimeout, true)
	if err != nil {
		return
	}

	return
}
