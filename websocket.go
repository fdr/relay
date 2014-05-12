package relay

import (
	"code.google.com/p/go.net/websocket"
	"github.com/fernet/fernet-go"
	"net/http"
	"time"
)

func WsDial(origin, url string, k *fernet.Key) (*FESession, error) {
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		return nil, err
	}

	s := NewFESession(*k)
	go s.Run(ws)

	return s, nil
}

type WsServer interface {
	KeySelect(*websocket.Conn) []*fernet.Key
	Handler(*BESession)
}

func WsHandler(s WsServer) http.Handler {
	hdlr := func(ws *websocket.Conn) {
		keys := s.KeySelect(ws)

		sess := NewBESession(keys, time.Minute*15)
		go sess.Run(ws)
		s.Handler(sess)
	}

	return &websocket.Server{
		Handler: hdlr,
	}
}
