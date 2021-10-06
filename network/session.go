package network

import (
	"net"

	log "github.com/sirupsen/logrus"
)

type Agent interface {
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SendMsg()
	ReadMsg()
	Close() error
}

type sendPack struct {
	data []byte
}

type recvPack struct {
	data []byte
}

type SessionKey uint64

func (s SessionKey) Parse() (areaId, typeId uint8, id uint16, sessionId uint32) {
	sessionId = uint32(s)
	areaId, typeId, id = ServerKey(s >> 32).Parse()
	return
}

// Session 连接会话
// 对应一个tcp/udp/websocket连接,对网络通信的封装，提供更多功能
// 1.应用层消息序列化与反序列化
// 2.将序列化数据分包发送或接收
//todo 3.中间件
type Session struct {
	ID        uint32
	SC        *ServiceConfig
	agent     Agent
	userData  map[interface{}]interface{}
	send      chan *sendPack // 消息发送队列
	recv      chan *recvPack // 消息接收队列
	closeSign chan struct{}
	pkgParser *PkgParser // 序列化数据分包发送或接收
	msgParser *MsgParser // 应用层消息序列化与反序列化
}

func NewSession(config *ServiceConfig) *Session {
	s := &Session{
		ID:        config.GetSeq(),
		SC:        config,
		userData:  make(map[interface{}]interface{}),
		send:      make(chan *sendPack, config.MaxSend),
		recv:      make(chan *recvPack, config.MaxRecv),
		closeSign: make(chan struct{}),
		pkgParser: gPkgParser,
		msgParser: gMsgParser,
	}
	return s
}

func (s *Session) Key() SessionKey {
	// 由低位到高位依次 SessionID(32位) ID(16位) Type(8位)  Area(8位)
	key := uint64(0)
	key |= uint64(s.SC.Key()) << 32
	key |= uint64(s.ID)
	return SessionKey(key)
}

func (s *Session) SetAgent(agent Agent) {
	s.agent = agent
}

func (s *Session) LocalAddr() net.Addr {
	return s.agent.LocalAddr()
}

func (s *Session) RemoteAddr() net.Addr {
	return s.agent.RemoteAddr()
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
		log.WithField("msgID", msgID).Errorf("send message error: %v", err)
		return
	}

	select {
	case <-s.closeSign:
		log.WithField("service", s.SC.String()).Trace("session closed")
	case s.send <- &sendPack{
		data: data,
	}:
	default:
		log.WithField("service", s.SC.String()).Error("close conn: channel full")
		s.Close()
	}
}

func (s *Session) sendMsg() {
	s.agent.SendMsg()
}

func (s *Session) readMsg() {
	s.agent.ReadMsg()
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

func (s *Session) Close() error {
	select {
	case <-s.closeSign:
		return nil
	default:
		close(s.closeSign)
	}

	err := s.agent.Close()

	select {
	case s.send <- nil:
	default:
	}
	return err
}
