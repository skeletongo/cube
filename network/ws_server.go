package network

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type WSServer struct {
	network   *Network
	SC        *ServiceConfig
	server    *http.Server
	ln        net.Listener
	upgrader  websocket.Upgrader
	sessions  map[*Session]struct{}
	sessionCh chan *Session
	connCh    chan *websocket.Conn
	closeSign chan struct{} // 触发服务关闭
	close     bool
}

func NewWSServer(network *Network, config *ServiceConfig) *WSServer {
	return &WSServer{
		network: network,
		SC:      config,
		upgrader: websocket.Upgrader{
			HandshakeTimeout: config.HTTPTimeout,
			CheckOrigin:      func(_ *http.Request) bool { return true },
			ReadBufferSize:   config.ReadBufferSize,
			WriteBufferSize:  config.WriteBufferSize,
		},
		sessions:  make(map[*Session]struct{}),
		sessionCh: make(chan *Session, 1000),
		connCh:    make(chan *websocket.Conn, 1000),
		closeSign: make(chan struct{}),
	}
}

func (w *WSServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(resp, "Method not allowed", 405)
		return
	}
	conn, err := w.upgrader.Upgrade(resp, req, nil)
	if err != nil {
		log.Warningf("upgrade error: %v", err)
		return
	}
	select {
	case w.connCh <- conn:
	default:
		conn.Close()
		log.WithField("ServiceInfo", w.SC).Error("connection channel full")
	}
}

func (w *WSServer) Start() error {
	addr := fmt.Sprintf("%s:%d", w.SC.Ip, w.SC.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.WithField("ServiceInfo", w.SC).Errorf("websocket server start error: %v", err)
		return err
	}
	log.WithField("ServiceInfo", w.SC).Trace("websocket server start")

	if w.SC.CertFile != "" || w.SC.KeyFile != "" {
		config := &tls.Config{}
		config.NextProtos = []string{"http/1.1"}

		var err error
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(w.SC.CertFile, w.SC.KeyFile)
		if err != nil {
			log.WithField("ServiceInfo", w.SC).Errorf("tls error: %v", err)
			return err
		}

		ln = tls.NewListener(ln, config)
	}

	w.ln = ln

	w.server = &http.Server{
		Addr:           addr,
		Handler:        w,
		ReadTimeout:    w.SC.HTTPTimeout,
		WriteTimeout:   w.SC.HTTPTimeout,
		MaxHeaderBytes: 1024,
	}

	go func() {
		defer func() { close(w.closeSign) }()
		if err = w.server.Serve(ln); err != nil {
			log.WithField("ServiceInfo", w.SC).Warningf("websocket httpServer error: %v", err)
			w.server.Close()
			w.ln.Close()
		}
	}()
	return nil
}

func (w *WSServer) Update() {
	for {
		select {
		case s := <-w.sessionCh:
			s.fireAfterClosed()
			delete(w.sessions, s)
			if w.close && len(w.sessions) == 0 {
				w.network.Release(w.SC)
				return
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
						w.network.Release(w.SC)
						return
					}
					break here
				}
			}

		case conn := <-w.connCh:
			if len(w.sessions) > w.SC.MaxConnNum {
				log.WithField("ServiceInfo", w.SC).Warning("too many connections")
				conn.Close()
				continue
			}

			var err error
			s := NewSession(w.SC)
			s.agent, err = NewWSSession(s, conn)
			if err != nil {
				log.WithField("ServiceInfo", w.SC).Errorf("NewWSSession error: %v", err)
				conn.Close()
				continue
			}

			w.sessions[s] = struct{}{}
			go s.sendMsg()
			go func() {
				s.readMsg()
				w.sessionCh <- s
			}()

			if !s.fireAfterConnected() {
				s.Close()
				continue
			}

		default:
			for v := range w.sessions {
				v.do()
			}
			return
		}
	}
}

func (w *WSServer) Shutdown() {
	log.WithField("ServiceInfo", w.SC).Trace("websocket server shutdown")
	w.server.Close()
	w.ln.Close()
}
