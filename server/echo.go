package server

import (
	"context"
	"io"
	"log"
	"net"
)

// An echo sever for test

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

func echoMain(ctx context.Context, addr string) {
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
	log.Println("Echo service listen on ", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("echo accpet failed: ", err)
			continue
		}
		go echo(ctx, conn)
	}
}
