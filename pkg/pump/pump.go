package pump

import (
	"context"
	"log"
	"net"
	"time"
)

const (
	MaxBuffSize    = 1024 * 1024
	DefaultTimeout = time.Second * 30
	ConnTimeout    = time.Microsecond * 200
)

type Pump struct {
	ctx      context.Context
	cancel   context.CancelFunc
	from, to net.Conn
	buff     []byte
	tmo      time.Duration
	timer    *time.Timer
}

func (p *Pump) Close() {
	p.from.Close()
	p.to.Close()
	p.cancel()
}

func isOK(err error) bool {
	neterr, ok := err.(net.Error)
	if ok && (neterr.Temporary() || neterr.Timeout()) {
		return true
	}
	return false
}

func (p *Pump) Run() {
	var offset, length, m, n int
	var err error
	for {
		select {
		case <-p.ctx.Done():
			log.Println("exit: ", p.ctx.Err().Error())
			return

		case <-p.timer.C:
			log.Println("Session timeout")
			p.Close()
			return

		default:
			if length < MaxBuffSize { // buffer is empty
				p.from.SetReadDeadline(time.Now().Add(ConnTimeout))
				m, err = p.from.Read(p.buff[length:])
				if !isOK(err) {
					log.Println("Read error: ", err.Error())
					p.Close()
					return
				}
				length += m
				p.timer.Reset(p.tmo)
			}
			if length > 0 { // some data is in buffer
				p.to.SetDeadline(time.Now().Add(ConnTimeout))
				n, err = p.to.Write(p.buff[offset:length])
				if !isOK(err) {
					log.Panicln("Write error: ", err.Error())
					p.Close()
					return
				}
				offset += n
				p.timer.Reset(p.tmo)
			}
			if offset == length { // all data has been send
				offset = 0
				length = 0
			}
		}
	}
}

func New(ctx context.Context, from, to net.Conn) (*Pump, error) {
	c, cancel := context.WithCancel(ctx)
	p := &Pump{
		ctx:    c,
		cancel: cancel,
		from:   from,
		to:     to,
		buff:   make([]byte, MaxBuffSize),
		tmo:    DefaultTimeout,
		timer:  time.NewTimer(DefaultTimeout),
	}
	go p.Run()
	return p, nil
}
