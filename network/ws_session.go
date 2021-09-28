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
	return w, nil
}

func (w *WSSession) SendMsg() {
	var zero time.Time
	for v := range w.Session.send {
		if v == nil {
			break
		}

		data, err := w.Session.pkgParser.Encode(v.data)
		if err != nil {
			log.Warningf("packet Encode error: %v", err)
			break
		}

		if w.Session.SC.WriteTimeout > 0 {
			w.Conn.SetWriteDeadline(time.Now().Add(w.Session.SC.WriteTimeout))
		}
		err = w.WriteMessage(websocket.BinaryMessage, data)
		w.Conn.SetWriteDeadline(zero)
		if err != nil {
			log.Warningf("websocket WriteMessage error: %v", err)
			break
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
	w.UnderlyingConn().(*net.TCPConn).SetLinger(0)
	return w.Conn.Close()
}
