package proxy

import (
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/gorilla/websocket"
)

type webSocketConn struct {
	r     *http.Request
	back  *websocket.Conn
	front *websocket.Conn
}

func (w *webSocketConn) Run(domain *Domain) {
	domain.WebSocketConnsLock.Lock()
	domain.WebSocketConns.Add(w)
	domain.WebSocketConnsLock.Unlock()

	defer func() {
		domain.WebSocketConnsLock.Lock()
		domain.WebSocketConns.Remove(w)
		domain.WebSocketConnsLock.Unlock()
	}()

	wait := make(chan bool, 4)
	go func() {
		defer func() {
			rec := recover()
			if rec != nil {
				logrus.WithFields(logrus.Fields{
					"panic": rec,
				}).Error("proxy: WebSocket back panic")
				wait <- true
			}
		}()
		io.Copy(w.back.UnderlyingConn(), w.front.UnderlyingConn())
		wait <- true
	}()
	go func() {
		defer func() {
			rec := recover()
			if rec != nil {
				logrus.WithFields(logrus.Fields{
					"panic": rec,
				}).Error("proxy: WebSocket front panic")
				wait <- true
			}
		}()
		io.Copy(w.front.UnderlyingConn(), w.back.UnderlyingConn())
		wait <- true
	}()
	<-wait

	w.Close()
}

func (w *webSocketConn) Close() {
	defer func() {
		recover()
	}()
	if w.back != nil {
		w.back.Close()
	}
	if w.front != nil {
		w.front.Close()
	}
}
