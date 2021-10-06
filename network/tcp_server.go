package network

import (
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
		log.WithField("service", t.SC).Errorf("tcp server start error: %v", err)
		return err
	}
	log.WithField("service", t.SC).Trace("tcp server start")

	t.ln = ln

	go func() {
		defer func() { close(t.closeSign) }()

		var tempDelay time.Duration
		for {
			conn, err := t.ln.Accept()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					if tempDelay == 0 {
						tempDelay = 5 * time.Millisecond
					} else {
						tempDelay *= 2
					}
					if max := 1 * time.Second; tempDelay > max {
						tempDelay = max
					}
					log.Warningf("accept error: %v; retrying in %v", err, tempDelay)
					time.Sleep(tempDelay)
					continue
				}
				log.WithField("service", t.SC).Warningf("tcp server listener error: %v", err)
				return
			}
			tempDelay = 0
			select {
			case t.connCh <- conn:
			default:
				conn.Close()
				log.Error("connection channel full")
			}
		}
	}()
	return nil
}

func (t *TCPServer) Update() {
	for {
		select {
		case s := <-t.sessionCh:
			delete(t.sessions, s)
			if t.close && len(t.sessions) == 0 {
				t.network.ServiceClosed(t.SC)
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
						t.network.ServiceClosed(t.SC)
						return
					}
					break here
				}
			}

		case conn := <-t.connCh:
			if len(t.sessions) > t.SC.MaxConnNum {
				log.Warning("too many connections")
				conn.Close()
				continue
			}

			s := NewSession(t.SC)
			agent, err := NewTCPSession(s, conn)
			if err != nil {
				log.WithField("service", t.SC).Errorf("NewTCPSession error: %v", err)
				conn.Close()
				continue
			}
			s.SetAgent(agent)
			t.sessions[s] = struct{}{}
			go s.sendMsg()
			go func() {
				s.readMsg()
				t.sessionCh <- s
			}()

		default:
			for v := range t.sessions {
				v.Do()
			}
			return
		}
	}
}

func (t *TCPServer) Shutdown() {
	log.WithField("service", t.SC).Trace("tcp server shutdown")
	t.ln.Close()
}
