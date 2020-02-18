package proxy

import (
	"net/http"
)

type ErrorHandler func(hand *Handler, rw http.ResponseWriter,
	r *http.Request, err error)
