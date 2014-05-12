package relay_test

import (
	"bytes"
	"github.com/deafbybeheading/femebe/core"
	"github.com/fdr/relay"
	"github.com/fernet/fernet-go"
	"strconv"
	"testing"
	"time"
)

func TestBEMessages(t *testing.T) {
	// Make a new key and start up a session
	k := fernet.Key{}
	k.Generate()
	sess := relay.NewBESession([]*fernet.Key{&k}, time.Minute*15)

	// Produce some messages that count upwards from zero
	//
	// Save a version that is not round-tripped through the
	// relay.Session as the expected results of a round trip.
	buf := &nopReadWriteCloser{}
	m := core.Message{}
	expected := &bytes.Buffer{}
	for i := 0; i < 30; i += 1 {
		ns := []byte(strconv.Itoa(i))

		expected.Write(ns)
		tok, err := fernet.EncryptAndSign(ns, &k)

		if err != nil {
			t.FailNow()
		}

		// Frame a message and put it in a buffer.
		m.InitFromBytes('F', tok)
		m.WriteTo(buf)
	}

	// Process the buffer.
	go sess.Run(buf)

	// Check the plaintext output.
	result := &bytes.Buffer{}
	sess.WriteTo(result)

	eb := expected.Bytes()
	rb := result.Bytes()

	if !bytes.Equal(eb, rb) {
		t.Fatalf("Round trip of bytes fails\n"+
			"expected:\n\t%s\nresults:\n\t%s", eb, rb)
	}
}
