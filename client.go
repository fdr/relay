package relay

import (
	"github.com/deafbybeheading/femebe/core"
	"github.com/fernet/fernet-go"
	"io"
)

type FESession struct {
	k fernet.Key

	rwc      io.ReadWriteCloser
	outbound chan []byte
	writeErr chan error
}

func NewFESession(key fernet.Key) *FESession {
	return &FESession{
		k:        key,
		outbound: make(chan []byte),
	}
}

func (s *FESession) Close() error {
	err := s.rwc.Close()
	close(s.outbound)
	return err
}

func (s *FESession) Write(p []byte) (n int, err error) {
	for {
		select {
		case s.outbound <- p:
			return len(p), nil
		case err := <-s.writeErr:
			return 0, err
		}
	}
}

func (s *FESession) Run(rwc io.ReadWriteCloser) {
	s.rwc = rwc

	cleanup := func(err error) {
		s.writeErr <- err
		close(s.writeErr)
	}

	m := core.Message{}
	for {
		buf, ok := <-s.outbound
		if !ok {
			return
		}

		tok, err := fernet.EncryptAndSign(buf, &s.k)
		if err != nil {
			cleanup(err)
			return
		}

		m.InitFromBytes('F', tok)
		_, err = m.WriteTo(rwc)

		if err != nil {
			cleanup(err)
			return
		}
	}

}
