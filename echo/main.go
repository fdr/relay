package main

import (
	"flag"
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

func (s *RelayEchoServer) KeySelect(req *http.Request) []*fernet.Key {
	return []*fernet.Key{testKey}
}

func (s *RelayEchoServer) Handler(be *relay.WsBESession) {
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
	fe, err := relay.WsDial("ws://localhost", "ws://localhost:12345",
		"thing", testKey)
	if err != nil {
		log.Fatalln("relay.WDial: " + err.Error())
	}

	io.Copy(fe, os.Stdin)
	return 0
}

func main() {
	flag.Parse()

	switch flag.Arg(0) {
	case "serve":
		os.Exit(serve(flag.Args()[1:]))
	case "client":
		os.Exit(client(flag.Args()[1:]))
	case "":
		log.Fatal("no subcommand specified.  " +
			"Specify \"serve\" or \"client\".")
	default:
		log.Fatalf("invalid subcommand: %v", flag.Args())
	}
}

func init() {
	testKey, _ = fernet.DecodeKey("15zbVcYNVyFV213OtajhDp3f+w10IvXpRsE/2OwRfSQ=")
}
