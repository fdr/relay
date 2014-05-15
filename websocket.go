package relay

import (
	"code.google.com/p/go.net/websocket"
	"errors"
	"github.com/fernet/fernet-go"
	"net/http"
	"time"
)

var ErrNoShake = errors.New("Could not handshake")

func WsDial(origin, url, what string, k *fernet.Key) (*FESession, error) {
	// Add mutual authentication headers (What, Verify) to
	// handshake.
	config, err := websocket.NewConfig(url, origin)
	if err != nil {
		return nil, err
	}

	config.Header.Add("What", what)
	tok, err := fernet.EncryptAndSign([]byte(what), k)
	if err != nil {
		return nil, err
	}

	config.Header.Add("Verify", string(tok))

	ws, err := websocket.DialConfig(config)
	if err != nil {
		return nil, err
	}

	s := NewFESession(*k)
	go s.Run(ws)

	return s, nil
}

type WsServer interface {
	KeySelect(*http.Request) []*fernet.Key
	Handler(*BESession)
}

func WsHandler(s WsServer) http.Handler {
	shake := func(c *websocket.Config, req *http.Request) error {
		keys := s.KeySelect(req)
		if keys == nil || len(keys) == 0 {
			return ErrNoShake
		}

		return nil
	}

	hdlr := func(ws *websocket.Conn) {
		// The handshake already checked this for the purpose
		// of fast-path rejection of the websocket upgrade,
		// but it's hard to stash the resultant selected
		// Fernet keys anywhere, so perform the same selection
		// again; it seems chaep enough.
		keys := s.KeySelect(ws.Request())
		if keys == nil || len(keys) == 0 {
			ws.Close()
			return
		}

		sess := NewBESession(keys, time.Minute*15)
		go sess.Run(ws)
		s.Handler(sess)
	}

	return &websocket.Server{
		Handshake: shake,
		Handler:   hdlr,
	}
}
