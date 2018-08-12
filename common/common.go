package common

import (
	"context"
	"net"

	"github.com/gorilla/websocket"
)

const (
	//BufferSize is size of read and write buffer
	BufferSize = 4 * 1048576 // 4M
)

// Ws2Tcp read data from websocket and write it to tcp
func Ws2Tcp(ctx context.Context, cancel context.CancelFunc, wsConn *websocket.Conn, tcpConn net.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := wsConn.ReadMessage()
			if err != nil {
				cancel()
			}
			var m int
			for m < len(message) {
				n, err := tcpConn.Write(message[m:])
				if err != nil {
					cancel()
					break
				}
				m += n
			}
		}
	}
}

// TCP2Ws read data from tcp and write it to websocket
func TCP2Ws(ctx context.Context, cancel context.CancelFunc, wsConn *websocket.Conn, tcpConn net.Conn) {
	buffer := make([]byte, BufferSize)
	for {
		select {
		case <-ctx.Done():
			wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		default:
			n, err := tcpConn.Read(buffer)
			if err != nil {
				cancel()
			} else {
				err = wsConn.WriteMessage(websocket.BinaryMessage, buffer[:n])
				if err != nil {
					cancel()
				}
			}
		}
	}
}
