package main

import (
	"code.google.com/p/go.net/websocket"
	"github.com/fdr/relay"
	"github.com/fernet/fernet-go"
	"io"
	"log"
	"net/http"
	"os"
)

var testKey *fernet.Key
var addr = "127.0.0.1:12345"

type RelayEchoServer struct{}

func (s *RelayEchoServer) KeySelect(ws *websocket.Conn) []*fernet.Key {
	return []*fernet.Key{testKey}
}

func (s *RelayEchoServer) Handler(be *relay.BESession) {
	be.WriteTo(os.Stdout)
}

func serve(args []string) int {
	http.Handle("/", relay.WsHandler(&RelayEchoServer{}))

	log.Println("starting server on", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalln("ListenAndServe: " + err.Error())
	}

	return 0
}

func client(args []string) int {
	fe, err := relay.WsDial("ws://localhost", "ws://localhost:12345", testKey)
	if err != nil {
		log.Fatalln("relay.WDial: " + err.Error())
	}

	io.Copy(fe, os.Stdin)
	return 0
}

func main() {
	switch os.Args[1] {
	case "serve":
		os.Exit(serve(os.Args[1:]))
	case "client":
		os.Exit(client(os.Args[1:]))
	default:
		log.Fatalf("invalid subcommand: %v", os.Args[1])
	}
}

func init() {
	testKey, _ = fernet.DecodeKey("15zbVcYNVyFV213OtajhDp3f+w10IvXpRsE/2OwRfSQ=")
}
