package network

import (
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

var TestTCPClientSession *Session

type TCPClient struct {
	network   *Network
	SC        *ServiceConfig
	dialCh    chan struct{} // 触发拨号
	sessions  map[*Session]struct{}
	sessionCh chan *Session
	connCh    chan net.Conn
	closeSign chan struct{} // 触发服务关闭
	dialSign  chan struct{} // 关闭拨号协程
	close     bool
}

func NewTCPClient(n *Network, config *ServiceConfig) *TCPClient {
	return &TCPClient{
		network:   n,
		SC:        config,
		dialCh:    make(chan struct{}, config.ClientNum),
		sessions:  make(map[*Session]struct{}, config.ClientNum),
		sessionCh: make(chan *Session, config.ClientNum),
		connCh:    make(chan net.Conn, config.ClientNum),
		closeSign: make(chan struct{}),
		dialSign:  make(chan struct{}),
	}
}

func (t *TCPClient) dial(addr string) net.Conn {
	for {
		conn, err := net.Dial("tcp", addr)
		select {
		case <-t.dialSign:
			if err == nil {
				conn.Close()
			}
			return nil
		default:
			if err == nil {
				return conn
			}
		}
		log.WithField("service", t.SC).Warningf("connect to %v error: %v", addr, err)
		time.Sleep(t.SC.ReconnectInterval)
	}
}

func (t *TCPClient) Start() error {
	addr := fmt.Sprintf("%s:%d", t.SC.Ip, t.SC.Port)
	for i := 0; i < t.SC.ClientNum; i++ {
		t.dialCh <- struct{}{}
	}
	log.WithField("service", t.SC).Trace("tcp client start")

	go func() {
		defer func() { close(t.closeSign) }()

		for {
			select {
			case <-t.dialSign:
				return
			case <-t.dialCh:
				conn := t.dial(addr)
				if conn == nil {
					continue
				}
				select {
				case t.connCh <- conn:
				default:
					log.Panicln("bug")
				}
			}
		}
	}()
	return nil
}

func (t *TCPClient) Update() {
	for {
		select {
		case s := <-t.sessionCh:
			delete(t.sessions, s)
			if t.close {
				if len(t.sessions) == 0 {
					t.network.ServiceClosed(t.SC)
					return
				}
				continue
			}
			// 重新拨号
			select {
			case t.dialCh <- struct{}{}:
			default:
				log.Panicln("bug")
			}

		case <-t.closeSign:
			if !t.close {
				t.close = true
				for v := range t.sessions {
					v.Close()
				}
				for {
					select {
					case conn := <-t.connCh:
						conn.Close()
					default:
						if len(t.sessions) == 0 {
							t.network.ServiceClosed(t.SC)
						}
						return
					}
				}
			}
			return

		case conn := <-t.connCh:
			s := NewSession(t.SC, NewTCPConn(conn, t.SC))
			t.sessions[s] = struct{}{}
			go s.sendMsg()
			go func() {
				s.readMsg()
				t.sessionCh <- s
			}()
			if TestTCPClientSession == nil {
				TestTCPClientSession = s
			}

		default:
			for s := range t.sessions {
				s.Do()
			}
			return
		}
	}
}

func (t *TCPClient) Shutdown() {
	log.WithField("service", t.SC).Trace("tcp client shutdown")
	close(t.dialSign)
}