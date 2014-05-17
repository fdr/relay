package relay_test

import (
	"bytes"
	"github.com/fdr/relay"
	"github.com/fernet/fernet-go"
	"strconv"
	"testing"
	"time"
)

func TestFEMessages(t *testing.T) {
	// Make a new key and start up a session
	k := fernet.Key{}
	k.Generate()
	sess := relay.NewFESession(k)

	// Drain bytes to buffer in a goroutine to prevent deadlock.
	buf := &nopReadWriteCloser{}
	go sess.Run(buf)

	// Write some messages that count upwards from zero
	for i := 0; i < 30; i += 1 {
		ns := []byte(strconv.Itoa(i))
		sess.Write(ns)
	}

	byts := buf.Bytes()

	if len(byts) == 0 {
		t.FailNow()
	}

	sess.Close()
}

func TestRoundTrip(t *testing.T) {
	k := fernet.Key{}
	k.Generate()

	// Set up a Front End session and a place for it to place
	// protocol traffic.
	buf := &nopReadWriteCloser{}
	feSess := relay.NewFESession(k)
	go feSess.Run(buf)

	expected := &bytes.Buffer{}

	// "Send" payloads.
	for i := 0; i < 30; i += 1 {
		ns := []byte(strconv.Itoa(i))
		expected.Write(ns)
		feSess.Write(ns)
	}

	// Start up a Backend session and de-frame those bytes.
	beSess := relay.NewBESession([]*fernet.Key{&k}, time.Minute*15)
	go beSess.Run(buf)
	result := &nopReadWriteCloser{}
	beSess.WriteTo(result)

	// Check that the round trip is symmetric.
	eb := expected.Bytes()
	rb := result.Bytes()

	if !bytes.Equal(eb, rb) {
		t.Fatalf("Round trip of bytes fails\n"+
			"expected:\n\t%s\nresults:\n\t%s", eb, rb)
	}

	feSess.Close()
	beSess.Close()
}
