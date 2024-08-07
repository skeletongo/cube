package network

import (
	"errors"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

type TCPServer struct {
	network   *Network
	SC        *ServiceConfig
	ln        net.Listener
	sessions  map[*Session]struct{}
	sessionCh chan *Session
	connCh    chan net.Conn
	closeSign chan struct{} // 触发服务关闭
	close     bool
}

func NewTCPServer(network *Network, config *ServiceConfig) *TCPServer {
	return &TCPServer{
		network:   network,
		SC:        config,
		sessions:  make(map[*Session]struct{}),
		sessionCh: make(chan *Session, 1000),
		connCh:    make(chan net.Conn, 1000),
		closeSign: make(chan struct{}),
	}
}

func (t *TCPServer) Start() error {
	addr := fmt.Sprintf("%s:%d", t.SC.Ip, t.SC.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.WithField("ServiceInfo", t.SC).Errorf("tcp server start error: %v", err)
		return err
	}
	log.WithField("ServiceInfo", t.SC).Trace("tcp server start")

	t.ln = ln

	go func() {
		defer func() { close(t.closeSign) }()

		var tempDelay time.Duration
		for {
			conn, err := t.ln.Accept()
			if err != nil {
				var ne net.Error
				if errors.As(err, &ne) && ne.Timeout() {
					if tempDelay == 0 {
						tempDelay = 5 * time.Millisecond
					} else {
						tempDelay *= 2
					}
					if duration := 1 * time.Second; tempDelay > duration {
						tempDelay = duration
					}
					log.WithField("ServiceInfo", t.SC).Warningf("accept error: %v; retrying in %v", err, tempDelay)
					time.Sleep(tempDelay)
					continue
				}
				log.WithField("ServiceInfo", t.SC).Warningf("tcp server listener error: %v", err)
				return
			}
			tempDelay = 0
			select {
			case t.connCh <- conn:
			default:
				conn.Close()
				log.WithField("ServiceInfo", t.SC).Error("connection channel full")
			}
		}
	}()
	return nil
}

func (t *TCPServer) Update() {
	for {
		select {
		case s := <-t.sessionCh:
			s.fireAfterClosed()
			delete(t.sessions, s)
			if t.close && len(t.sessions) == 0 {
				t.network.Release(t.SC)
				return
			}

		case <-t.closeSign:
			t.closeSign = make(chan struct{})
			t.close = true
			for v := range t.sessions {
				v.Close()
			}
		here:
			for {
				select {
				case conn := <-t.connCh:
					conn.Close()
				default:
					if len(t.sessions) == 0 {
						t.network.Release(t.SC)
						return
					}
					break here
				}
			}

		case conn := <-t.connCh:
			if len(t.sessions) > t.SC.MaxConnNum {
				log.WithField("ServiceInfo", t.SC).Warning("too many connections")
				conn.Close()
				continue
			}

			var err error
			s := NewSession(t.SC)
			s.agent, err = NewTCPSession(s, conn)
			if err != nil {
				log.WithField("ServiceInfo", t.SC).Errorf("NewTCPSession error: %v", err)
				conn.Close()
				continue
			}

			t.sessions[s] = struct{}{}
			go s.sendMsg()
			go func() {
				s.readMsg()
				t.sessionCh <- s
			}()

			if !s.fireAfterConnected() {
				s.Close()
				continue
			}

		default:
			for v := range t.sessions {
				v.do()
			}
			return
		}
	}
}

func (t *TCPServer) Shutdown() {
	log.WithField("ServiceInfo", t.SC).Trace("tcp server shutdown")
	t.ln.Close()
}
