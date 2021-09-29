package network

import (
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type WSClient struct {
	network   *Network
	SC        *ServiceConfig
	dialer    websocket.Dialer
	dialCh    chan struct{} // 触发拨号
	sessions  map[*Session]struct{}
	sessionCh chan *Session
	connCh    chan *websocket.Conn
	closeSign chan struct{} // 触发服务关闭
	dialSign  chan struct{} // 关闭拨号协程
	close     bool
}

func NewWSClient(n *Network, config *ServiceConfig) *WSClient {
	return &WSClient{
		network: n,
		SC:      config,
		dialer: websocket.Dialer{
			HandshakeTimeout: config.HTTPTimeout,
			ReadBufferSize:   config.ReadBufferSize,
			WriteBufferSize:  config.WriteBufferSize,
		},
		dialCh:    make(chan struct{}, config.ClientNum),
		sessions:  make(map[*Session]struct{}, config.ClientNum),
		sessionCh: make(chan *Session, config.ClientNum),
		connCh:    make(chan *websocket.Conn, config.ClientNum),
		closeSign: make(chan struct{}),
		dialSign:  make(chan struct{}),
	}
}

func (w *WSClient) dial(url string) *websocket.Conn {
	for {
		conn, _, err := w.dialer.Dial(url, nil)
		select {
		case <-w.dialSign:
			if err == nil {
				conn.Close()
			}
			return nil
		default:
			if err == nil {
				return conn
			}
		}
		log.WithField("service", w.SC).Warningf("connect to %v error: %v", url, err)
		time.Sleep(w.SC.ReconnectInterval)
	}
}

func (w *WSClient) Start() error {
	urlStr := w.SC.Protocol + "://" + w.SC.Ip + ":" + strconv.Itoa(int(w.SC.Port)) + w.SC.Path
	for i := 0; i < w.SC.ClientNum; i++ {
		w.dialCh <- struct{}{}
	}
	log.WithField("service", w.SC).Trace("websocket client start")

	go func() {
		defer func() { close(w.closeSign) }()

		for {
			select {
			case <-w.dialSign:
				return
			case <-w.dialCh:
				conn := w.dial(urlStr)
				if conn == nil {
					continue
				}
				select {
				case w.connCh <- conn:
				default:
					log.Panicln("bug")
				}
			}
		}
	}()
	return nil
}

func (w *WSClient) Update() {
	for {
		select {
		case s := <-w.sessionCh:
			delete(w.sessions, s)
			if w.close {
				if len(w.sessions) == 0 {
					w.network.ServiceClosed(w.SC)
					return
				}
				continue
			}
			// 重新拨号
			select {
			case w.dialCh <- struct{}{}:
			default:
				log.Panicln("bug")
			}

		case <-w.closeSign:
			w.closeSign = make(chan struct{})
			w.close = true
			for v := range w.sessions {
				v.Close()
			}
		here:
			for {
				select {
				case conn := <-w.connCh:
					conn.Close()
				default:
					if len(w.sessions) == 0 {
						w.network.ServiceClosed(w.SC)
						return
					}
					break here
				}
			}

		case conn := <-w.connCh:
			var err error
			s := NewSession(w.SC)
			s.Agent, err = NewWSSession(s, conn)
			if err != nil {
				log.WithField("service", w.SC).Errorf("NewWSSession error: %v", err)
				conn.Close()
				continue
			}
			w.sessions[s] = struct{}{}
			go s.sendMsg()
			go func() {
				s.readMsg()
				w.sessionCh <- s
			}()
			// test
			//type Ping struct {
			//	Data string
			//}
			//s.Send(1, &Ping{Data: "ping"})
			// test

		default:
			for s := range w.sessions {
				s.Do()
			}
			return
		}
	}
}

func (w *WSClient) Shutdown() {
	log.WithField("service", w.SC).Trace("websocket client shutdown")
	close(w.dialSign)
}
