package proxy

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
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

		for {
			msgType, msg, err := w.front.ReadMessage()
			if err != nil {
				closeMsg := websocket.FormatCloseMessage(
					websocket.CloseNormalClosure, fmt.Sprintf("%v", err))
				if e, ok := err.(*websocket.CloseError); ok {
					if e.Code != websocket.CloseNoStatusReceived {
						closeMsg = websocket.FormatCloseMessage(e.Code, e.Text)
					}
				}
				_ = w.back.WriteMessage(websocket.CloseMessage, closeMsg)
				break
			}

			_ = w.back.WriteMessage(msgType, msg)
		}

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

		for {
			msgType, msg, err := w.back.ReadMessage()
			if err != nil {
				closeMsg := websocket.FormatCloseMessage(
					websocket.CloseNormalClosure, fmt.Sprintf("%v", err))
				if e, ok := err.(*websocket.CloseError); ok {
					if e.Code != websocket.CloseNoStatusReceived {
						closeMsg = websocket.FormatCloseMessage(e.Code, e.Text)
					}
				}
				_ = w.front.WriteMessage(websocket.CloseMessage, closeMsg)
				break
			}

			_ = w.front.WriteMessage(msgType, msg)
		}

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
