package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/fanzy618/wsproxy/common"

	"github.com/gorilla/websocket"
)

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
	conn := wsConn.UnderlyingConn()
	if c, ok := conn.(*net.TCPConn); ok {
		c.SetReadBuffer(common.BufferSize)
		c.SetWriteBuffer(common.BufferSize)
		c.SetKeepAlive(true)
		c.SetKeepAlivePeriod(30 * time.Second)
	}
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

// Config is where configuration of client
type Config struct {
	LocalAddr       string
	RemoteAddr      string
	ServerAddr      string
	InteractiveMode bool
	SkipVerify      bool
	RootCA          string
	Cert            string
	Key             string
}

// Main is the entry point of client
func Main(ctx context.Context, cfg Config) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctx = context.WithValue(ctx, ServiceAddr, cfg.ServerAddr)
	ctx = context.WithValue(ctx, RemoteAddr, cfg.RemoteAddr)

	lc := new(net.ListenConfig)
	listener, err := lc.Listen(ctx, "tcp4", cfg.LocalAddr)
	if err != nil {
		log.Printf("Listen on %s failed:%s\n", cfg.LocalAddr, err)
		return
	}
	defer listener.Close()
	log.Println("Service listen on ", cfg.LocalAddr)

	if cfg.RootCA != "" {
		var caCertPool *x509.CertPool
		if cfg.RootCA != "" {
			clientCA, err := ioutil.ReadFile(cfg.RootCA)
			if err != nil {
				log.Fatal(err)
			}
			caCertPool = x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(clientCA)
		}
		clientCert, err := tls.LoadX509KeyPair(cfg.Cert, cfg.Key)
		if err != nil {
			log.Printf("Listen on %s failed:%s\n", cfg.LocalAddr, err)
			return
		}
		dialer.TLSClientConfig = &tls.Config{
			RootCAs:            caCertPool,
			InsecureSkipVerify: cfg.SkipVerify,
			Certificates:       []tls.Certificate{clientCert},
		}
	}

mainloop:
	for {
		conn, err := listener.Accept()
		select {
		case <-ctx.Done():
			log.Println("Client exist because ", ctx.Err())
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
			c.SetKeepAlive(true)
			c.SetKeepAlivePeriod(30 * time.Second)
		}
		go proxy(ctx, conn)
	}

	log.Println("Service exit")
}
