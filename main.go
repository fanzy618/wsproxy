package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"wsproxy/client"
	"wsproxy/server"
)

// client side flags
var l = flag.String("l", "0.0.0.0:5004", "Local address")
var r = flag.String("r", "127.0.0.1:3128", "Remote address")
var s = flag.String("s", "ws://127.0.0.1:1443/proxy", "Server's address")
var i = flag.Bool("i", false, "Use stdin as input and write output to stdout")

// server side flags
var a = flag.String("a", "0.0.0.0:1443", "websocket service address")

var role = flag.String("role", "server", "server | client")

func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.SetOutput(os.Stderr)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	switch *role {
	case "server":
		cfg := server.Config{
			WebSocketAddr: *a,
		}
		go server.Main(ctx, cfg)
	case "client":
		cfg := client.Config{
			LocalAddr:       *l,
			RemoteAddr:      *r,
			ServerAddr:      *s,
			InteractiveMode: *i,
		}
		go client.Main(ctx, cfg)
	default:
		log.Fatal("Unknown role ", *role)
	}

	log.Println("Starting service as a wsproxy", *role)
	s := <-sc
	cancel()
	log.Println("Get signal:", s)
}
