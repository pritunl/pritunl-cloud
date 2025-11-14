package dnss

import (
	"github.com/miekg/dns"
)

type Response struct {
	dns.ResponseWriter
	msg *dns.Msg
}

func (r *Response) WriteMsg(m *dns.Msg) error {
	r.msg = m
	return nil
}
