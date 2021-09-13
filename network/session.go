package network

import (
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

type sendPack struct {
	data [][]byte
}

type recvPack struct {
	data []byte
}

// Session 连接会话
// 对应一个tcp/udp/websocket连接,对网络通信的封装，提供更多功能
// 1.应用层消息序列化与反序列化
// 2.将序列化数据分包发送或接收
//todo 3.中间件
type Session struct {
	ID        int
	SC        *ServiceConfig
	userData  map[interface{}]interface{}
	send      chan *sendPack // 消息发送队列
	recv      chan *recvPack // 消息接收队列
	closeSign chan struct{}
	pkgParser *PkgParser // 序列化数据分包发送或接收
	msgParser *MsgParser // 应用层消息序列化与反序列化
	conn      Conn
}

func NewSession(config *ServiceConfig, conn Conn) *Session {
	s := &Session{
		ID:        config.GetSeq(),
		SC:        config,
		userData:  make(map[interface{}]interface{}),
		send:      make(chan *sendPack, config.MaxSend),
		recv:      make(chan *recvPack, config.MaxRecv),
		closeSign: make(chan struct{}),
		pkgParser: gPkgParser,
		msgParser: gMsgParser,
		conn:      conn,
	}
	return s
}

func (s *Session) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

func (s *Session) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

func (s *Session) SetData(key, value interface{}) {
	s.userData[key] = value
}

func (s *Session) GetData(key interface{}) interface{} {
	return s.userData[key]
}

func (s *Session) HasData(key interface{}) bool {
	_, ok := s.userData[key]
	return ok
}

func (s *Session) DelData(key interface{}) {
	delete(s.userData, key)
}

func (s *Session) Send(msgID uint16, msg interface{}) {
	data, err := s.msgParser.Marshal(msgID, msg)
	if err != nil {
		log.WithField("msgID", msgID).Error("send message error: %v", err)
		return
	}

	select {
	case <-s.closeSign:
		log.WithField("msgID", msgID).Trace("session closed")
	case s.send <- &sendPack{
		data: data,
	}:
	default:
		log.WithField("msgID", msgID).Error("close conn: channel full")
		s.Close()
	}
}

// send goroutine
func (s *Session) sendMsg() {
	var zero time.Time
	for v := range s.send {
		if v == nil {
			break
		}
		if s.SC.WriteTimeout > 0 {
			s.conn.SetWriteDeadline(time.Now().Add(s.SC.WriteTimeout))
		}
		err := s.pkgParser.Encode(s.conn, v.data)
		s.conn.SetWriteDeadline(zero)
		if err != nil {
			log.Warningf("packet encode error: %v", err)
			break
		}
	}

	s.Close()
}

// read goroutine
func (s *Session) readMsg() {
	var zero time.Time
	for {
		if s.SC.ReadTimeout > 0 {
			s.conn.SetReadDeadline(time.Now().Add(s.SC.ReadTimeout))
		}
		data, err := s.pkgParser.Decode(s.conn)
		s.conn.SetReadDeadline(zero)
		if err != nil {
			log.Warningf("packet decode error: %v", err)
			break
		}
		r := &recvPack{
			data: data,
		}
		s.recv <- r
	}

	s.Close()
}

// goroutine safe
func (s *Session) Close() {
	select {
	case <-s.closeSign:
		return
	default:
		close(s.closeSign)
	}
	s.conn.Close()

	select {
	case s.send <- nil:
	default:
	}
}

func (s *Session) Do() {
	for i := 0; i < s.SC.MaxRecv; i++ {
		select {
		case v := <-s.recv:
			msgID, msg, err := s.msgParser.Unmarshal(v.data)
			if err != nil {
				log.Errorf("message unmarshal error: %v", err)
				continue
			}

			h := GetHandler(msgID)
			if h == nil {
				log.WithField("msgID", msgID).Warning("protocol number not register")
				return
			}
			if err := h.Process(s, msgID, msg); err != nil {
				log.WithField("msgID", msgID).Errorf("process error: %v", err)
			}
		default:
			return
		}
	}
}
