package server

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/fanzy618/wsproxy/common"
)

func echo(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := io.Copy(conn, conn)
			if err != nil {
				log.Println("echo: ", err)
				return
			}
		}
	}
}

func echoMain(addr string) {
	laddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		log.Printf("Echo listen on %s failed:%s\n", addr, err)
		return
	}

	listener, err := net.ListenTCP("tcp4", laddr)
	if err != nil {
		log.Printf("Listen on %s failed:%s\n", addr, err)
		return
	}
	defer listener.Close()
	log.Println("Service listen on ", addr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("echo accpet failed: ", err)
			continue
		}
		go echo(ctx, conn)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  common.BufferSize,
	WriteBufferSize: common.BufferSize,
}

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
}

// Main is the entry point of server
func Main(ctx context.Context, cfg Config) {
	http.HandleFunc("/proxy", Proxy)
	log.Fatalln(http.ListenAndServe(cfg.WebSocketAddr, nil))
}
