package proxy

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
)

func WriteError(w http.ResponseWriter, r *http.Request, code int, err error) {
	http.Error(w, utils.GetStatusMessage(code), code)

	logrus.WithFields(logrus.Fields{
		"client": node.Self.GetRemoteAddr(r),
		"error":  err,
	}).Error("proxy: Serve error")
}
