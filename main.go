package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fanzy618/wsproxy/client"
	"github.com/fanzy618/wsproxy/server"
)

// common flags
var role = flag.String("role", "server", "server | client")
var rootCA = flag.String("root-ca", "", "Root CA file")

// client side flags
var l = flag.String("l", "0.0.0.0:80", "Local address")
var r = flag.String("r", "127.0.0.1:3128", "Remote address")
var s = flag.String("s", "ws://127.0.0.1:1443/proxy", "Server's address")
var i = flag.Bool("i", false, "Use stdin as input and write output to stdout")
var skipVerify = flag.Bool("k", false, "Insecury skip server ca verify")

// server side flags
var a = flag.String("a", "0.0.0.0:443", "websocket service address")
var serverKey = flag.String("server-key", "", "Server's CA key")
var serverCA = flag.String("server-ca", "", "Server's CA file")
var proxy = flag.Bool("proxy", false, "Enable a proxy server")

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
			ServerKey:     *serverKey,
			ServerCA:      *serverCA,
			RootCA:        *rootCA,
			ProxyEnable: *proxy,
		}
		go server.Main(ctx, cfg)
	case "client":
		cfg := client.Config{
			LocalAddr:  *l,
			RemoteAddr: *r,
			ServerAddr: *s,
			SkipVerify: *skipVerify,
			RootCA:     *rootCA,
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
