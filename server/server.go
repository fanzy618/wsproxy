package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/fanzy618/wsproxy/common"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  common.BufferSize,
	WriteBufferSize: common.BufferSize,
}

const echoAddr = "127.0.0.1:11234"

// Proxy handle URL like "/proxy?des=127.0.0.1:3128"
func Proxy(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		w.WriteHeader(500)
		return
	}
	defer c.Close()

	log.Println("Handle request: ", r.URL.String())
	defer log.Println("Close request ", r.URL.String())

	dst := r.URL.Query().Get("dst")
	if dst == "echo" || dst == "" {
		dst = echoAddr
	}

	addr, err := net.ResolveTCPAddr("tcp4", dst)
	if err != nil {
		log.Println("Resolve ", dst, " failed: ", err.Error())
		w.WriteHeader(500)
		return
	}
	conn, err := net.DialTCP("tcp4", nil, addr)
	if err != nil {
		log.Println("Connect to ", dst, " failed: ", err.Error())
		w.WriteHeader(500)
		return
	}
	defer conn.Close()

	conn.SetReadBuffer(common.BufferSize)
	conn.SetWriteBuffer(common.BufferSize)
	log.Println("Connect to ", dst, " succeed!")

	ctx, cancel := context.WithCancel(context.Background())
	go common.Ws2Tcp(ctx, cancel, c, conn)
	go common.TCP2Ws(ctx, cancel, c, conn)

	for {
		select {
		case <-ctx.Done():
			return
		}
	}
}

// Config for server
type Config struct {
	WebSocketAddr string
	QuicAddr      string
	ServerKey     string
	ServerCA      string
	RootCA        string

	ProxyEnable bool
	EchoEnable  bool
}

// OK return string "OK" for test and health check
func OK(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// Main is the entry point of server
func Main(ctx context.Context, cfg Config) {
	if cfg.EchoEnable {
		go echoMain(ctx, echoAddr)
	}

	http.HandleFunc("/proxy", Proxy)
	http.HandleFunc("/", OK)

	if cfg.ServerCA != "" && cfg.ServerKey != "" {
		var server *http.Server
		var tlsCfg *tls.Config
		if cfg.RootCA != "" {
			clientCA, err := ioutil.ReadFile(cfg.RootCA)
			if err != nil {
				log.Fatal(err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(clientCA)
			tlsCfg = &tls.Config{
				ClientAuth: tls.RequireAndVerifyClientCert,
				ClientCAs:  caCertPool,
			}
		}
		addr := cfg.WebSocketAddr
		if addr == "" {
			addr = ":443"
		}
		server = &http.Server{
			Addr:      cfg.WebSocketAddr,
			TLSConfig: tlsCfg,
		}
		log.Fatalln(server.ListenAndServeTLS(cfg.ServerCA, cfg.ServerKey))
		return
	}

	log.Fatalln(http.ListenAndServe(cfg.WebSocketAddr, nil))
}
