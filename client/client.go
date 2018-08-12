package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wsproxy/common"

	"github.com/gorilla/websocket"
)

var l = flag.String("l", "0.0.0.0:5004", "Local address")
var r = flag.String("r", "127.0.0.1:3128", "Remote address")
var s = flag.String("s", "ws://127.0.0.1:1443/proxy", "Server's address")

// ProxyParement is the type of key in context
type ProxyParement string

const (
	//ServiceAddr is the key of service addr
	ServiceAddr ProxyParement = "ServiceAddress"
	//RemoteAddr is as it name
	RemoteAddr ProxyParement = "RemoteAddress"
)

var dialer = &websocket.Dialer{
	ReadBufferSize:  common.BufferSize,
	WriteBufferSize: common.BufferSize,
}

func getURL(ctx context.Context) string {
	sa, ok := ctx.Value(ServiceAddr).(string)
	if !ok || sa == "" {
		return ""
	}

	dst, ok := ctx.Value(RemoteAddr).(string)
	if !ok || dst == "" {
		return ""
	}

	return fmt.Sprintf("%s?dst=%s", sa, dst)
}

func proxy(ctx context.Context, tcpConn net.Conn) {
	defer tcpConn.Close()

	log.Println("Handle connection from ", tcpConn.RemoteAddr())
	defer log.Println("Close connection from ", tcpConn.RemoteAddr())

	url := getURL(ctx)
	if url == "" {
		log.Println("Get URL failed.")
		return
	}

	wsConn, resp, err := dialer.Dial(url, nil)
	if err != nil {
		log.Printf("Connect to %s failed: %s. Response is %+v.",
			url, err, resp)
		return
	}
	defer wsConn.Close()
	ctx, cancel := context.WithCancel(ctx)

	go common.Ws2Tcp(ctx, cancel, wsConn, tcpConn)
	go common.TCP2Ws(ctx, cancel, wsConn, tcpConn)

	for {
		select {
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

	laddr, err := net.ResolveTCPAddr("tcp4", *l)
	if err != nil {
		log.Printf("Listen on %s failed:%s\n", *l, err)
		return
	}

	listener, err := net.ListenTCP("tcp4", laddr)
	if err != nil {
		log.Printf("Listen on %s failed:%s\n", *l, err)
		return
	}
	defer listener.Close()
	log.Println("Service listen on ", *l)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = context.WithValue(ctx, ServiceAddr, *s)
	ctx = context.WithValue(ctx, RemoteAddr, *r)
mainloop:
	for {
		listener.SetDeadline(time.Now().Add(time.Second))
		conn, err := listener.Accept()
		select {
		case s := <-sc:
			log.Println("Get signal ", s)
			break mainloop
		default:
			// make it non-block
		}
		if err != nil {
			e, ok := err.(net.Error)
			if !ok || !(e.Timeout() && e.Temporary()) {
				log.Println("Accept failed: ", err.Error())
			}
			continue
		}
		if c, ok := conn.(*net.TCPConn); ok {
			c.SetReadBuffer(common.BufferSize)
			c.SetWriteBuffer(common.BufferSize)
		}
		go proxy(ctx, conn)
	}

	log.Println("Service exit")
}
