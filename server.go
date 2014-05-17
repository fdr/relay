package relay

import (
	"github.com/deafbybeheading/femebe/core"
	"github.com/fernet/fernet-go"
	"io"
	"log"
	"time"
)

type IngressError struct {
	error
}

type EgressError struct {
	error
}

type ioPair struct {
	buf []byte
	err error
}

type BESessionDecoder interface {
	io.WriterTo
	io.Closer
	Run(rwc io.ReadWriteCloser)
	Keys() []*fernet.Key
}

type BESession struct {
	k   []*fernet.Key
	ttl time.Duration
	rwc io.ReadWriteCloser

	packets chan *ioPair
}

func NewBESession(keys []*fernet.Key, ttl time.Duration) *BESession {

	return &BESession{
		k:       keys,
		ttl:     ttl,
		packets: make(chan *ioPair),
	}
}

func (s *BESession) Close() error {
	err := s.rwc.Close()
	close(s.packets)
	return err
}

func (s *BESession) Keys() []*fernet.Key {
	return s.k
}

func (s *BESession) WriteTo(w io.Writer) (n int64, err error) {
	for {
		p, ok := <-s.packets
		if !ok {
			return n, nil
		}

		wrote, err := w.Write(p.buf)

		n += int64(wrote)
		if err != nil {
			return n, EgressError{err}
		}

		if p.err != nil {
			if p.err == io.EOF {
				return n, nil
			}

			return n, IngressError{p.err}
		}
	}
}

func (s *BESession) Run(rwc io.ReadWriteCloser) {
	s.rwc = rwc
	c := core.NewBackendStream(rwc)
	m := core.Message{}

	for {
		err := c.Next(&m)

		if err != nil {
			s.packets <- &ioPair{err: err}
			return
		}

		switch m.MsgType() {
		case 'F':
			// Fernet Packet
			tok, err := m.Force()
			if err != nil {
				s.packets <- &ioPair{err: err}
				close(s.packets)
			}

			plaintext := fernet.VerifyAndDecrypt(tok, time.Minute*15, s.k)
			if plaintext == nil {
				log.Println("fernet message failed verification")
			} else {
				s.packets <- &ioPair{buf: plaintext}
			}
		}
	}
}
