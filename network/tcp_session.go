package network

import (
	"bytes"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

type TCPSession struct {
	net.Conn
	Session *Session
}

func NewTCPSession(s *Session, conn net.Conn) (*TCPSession, error) {
	t := &TCPSession{
		Conn:    conn,
		Session: s,
	}

	var err error
	c := conn.(*net.TCPConn)
	if s.SC.Linger > 0 {
		if err = c.SetLinger(s.SC.Linger); err != nil {
			return nil, err
		}
	}
	if err = c.SetKeepAlive(s.SC.KeepAlive); err != nil {
		return nil, err
	}
	if s.SC.KeepAlive && s.SC.KeepAlivePeriod > 0 {
		if err = c.SetKeepAlivePeriod(s.SC.KeepAlivePeriod); err != nil {
			return nil, err
		}
	}
	if s.SC.ReadBufferSize > 0 {
		if err = c.SetReadBuffer(s.SC.ReadBufferSize); err != nil {
			return nil, err
		}
	}
	if s.SC.WriteBufferSize > 0 {
		if err = c.SetWriteBuffer(s.SC.WriteBufferSize); err != nil {
			return nil, err
		}
	}
	return t, err
}

func (t *TCPSession) SendMsg() {
	var zero time.Time
	for v := range t.Session.send {
		if v == nil {
			break
		}

		if t.Session.SC.WriteTimeout > 0 {
			t.Conn.SetWriteDeadline(time.Now().Add(t.Session.SC.WriteTimeout))
		}
		err := gPkgParser.EncodeByWriter(t.Conn, v.data)
		t.Conn.SetWriteDeadline(zero)
		if err != nil {
			log.Warningf("TCP EncodeByWriter error: %v", err)
			putBuffer(bytes.NewBuffer(v.data))
			break
		}

		// 解码后再由过滤器处理
		t.Session.fireSendMsgAfterSend(v)
	}

	t.Session.Close()
}

func (t *TCPSession) ReadMsg() {
	var zero time.Time
	for {
		if t.Session.SC.ReadTimeout > 0 {
			t.Conn.SetReadDeadline(time.Now().Add(t.Session.SC.ReadTimeout))
		}
		data, err := gPkgParser.DecodeByReader(t.Conn)
		t.Conn.SetReadDeadline(zero)
		if err != nil {
			log.Warningf("TCP DecodeByReader error: %v", err)
			break
		}

		t.Session.recv <- data
	}

	t.Session.Close()
}

func (t *TCPSession) Close() error {
	return t.Conn.Close()
}
