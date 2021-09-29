package network

import (
	"io"
	"net"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type WSSession struct {
	*websocket.Conn
	Session *Session
}

func NewWSSession(s *Session, conn *websocket.Conn) (*WSSession, error) {
	w := &WSSession{
		Conn:    conn,
		Session: s,
	}
	conn.SetReadLimit(int64(Config.LenMsgLen + Config.MaxMsgLen))

	var err error
	c := conn.UnderlyingConn().(*net.TCPConn)
	if s.SC.Linger > 0 {
		if err = c.SetLinger(s.SC.Linger); err != nil {
			return nil, err
		}
	}
	return w, nil
}

func (w *WSSession) SendMsg() {
	var err error
	var zero time.Time
	var writer io.WriteCloser
here:
	for {
		if writer != nil {
			writer.Close()
			writer = nil
		}

		select {
		case v := <-w.Session.send:
			if v == nil {
				break here
			}

			if writer, err = w.NextWriter(websocket.BinaryMessage); err != nil {
				log.Warningf("websocket NextWriter error: %v", err)
				break
			}

			if w.Session.SC.WriteTimeout > 0 {
				w.Conn.SetWriteDeadline(time.Now().Add(w.Session.SC.WriteTimeout))
			}
			err = w.Session.pkgParser.EncodeByIOWriter(writer, v.data)
			w.Conn.SetWriteDeadline(zero)
			if err != nil {
				log.Warningf("websocket WriteMessage error: %v", err)
				break
			}
		}
	}

	w.Session.Close()
}

func (w *WSSession) ReadMsg() {
	var err error
	var zero time.Time
	var reader io.Reader
	for {
		if _, reader, err = w.NextReader(); err != nil {
			log.Warningf("websocket NextReader error: %v", err)
			break
		}

		if w.Session.SC.ReadTimeout > 0 {
			w.Conn.SetReadDeadline(time.Now().Add(w.Session.SC.ReadTimeout))
		}
		data, err := w.Session.pkgParser.DecodeByIOReader(reader)
		w.Conn.SetReadDeadline(zero)
		if err != nil {
			log.Warningf("packet DecodeByIOReader error: %v", err)
			break
		}

		r := &recvPack{
			data: data,
		}
		w.Session.recv <- r
	}

	w.Session.Close()
}

func (w *WSSession) Close() error {
	return w.Conn.Close()
}
