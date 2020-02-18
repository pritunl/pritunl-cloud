package proxy

import (
	"net/http"
)

type RespHandler func(hand *Handler, resp *http.Response) (err error)

type ErrorHandler func(hand *Handler, rw http.ResponseWriter,
	r *http.Request, err error)
