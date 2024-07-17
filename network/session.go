package network

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/base"
	"github.com/skeletongo/cube/module"
)

// Agent 连接
type Agent interface {
	// LocalAddr 获取本地地址
	LocalAddr() net.Addr

	// RemoteAddr 获取远程地址
	RemoteAddr() net.Addr

	// SendMsg 发送消息线程
	SendMsg()

	// ReadMsg 读取客户端消息线程
	ReadMsg()

	// Close 连接关闭方法
	Close() error
}

type sendPack struct {
	data    []byte
	msgType reflect.Type
}

// SessionKey 连接标识
type SessionKey uint64

func (s SessionKey) Parse() (areaId, typeId uint8, id uint16, sessionId uint32) {
	sessionId = uint32(s)
	areaId, typeId, id = ServerKey(s >> 32).Parse()
	return
}

// Session 连接会话
// 对应一个tcp/websocket连接,对网络通信的封装，提供更多功能
type Session struct {
	ID        uint32
	SC        *ServiceConfig
	context   *Context
	agent     Agent          // 连接实例
	send      chan *sendPack // 消息发送队列
	recv      chan []byte    // 消息接收队列
	closeSign chan struct{}
}

func NewSession(config *ServiceConfig) *Session {
	s := &Session{
		ID:        config.getSeq(),
		SC:        config,
		send:      make(chan *sendPack, config.MaxSend),
		recv:      make(chan []byte, config.MaxRecv),
		closeSign: make(chan struct{}),
	}
	s.context = &Context{
		Session: s,
		Keys:    sync.Map{},
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

func (s *Session) LocalAddr() net.Addr {
	return s.agent.LocalAddr()
}

func (s *Session) RemoteAddr() net.Addr {
	return s.agent.RemoteAddr()
}

// Send 发送消息
// msgID 消息号
// msg 消息数据
// 线程不安全，必须在module节点上执行
func (s *Session) Send(msgID uint16, msg interface{}) {
	// update context
	s.context.MsgID = msgID
	s.context.Msg = msg
	if !s.fireBeforeSend() {
		return
	}

	msgType := reflect.TypeOf(s.context.Msg)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		log.WithField("msgID", s.context.MsgID).Error("message pointer required")
		return
	}

	data, err := gMsgParser.Marshal(s.context.MsgID, s.context.Msg, int(Config.LenMsgLen))
	if err != nil {
		log.WithField("msgID", s.context.MsgID).Errorf("send message error: %v", err)
		return
	}

	select {
	case <-s.closeSign:
		log.WithField("SessionInfo", s).Trace("session closed")
	case s.send <- &sendPack{data: data, msgType: msgType}:
	default:
		log.WithField("SessionInfo", s).Error("close conn: channel full")
		_ = s.Close()
	}
}

func (s *Session) fireAfterConnected() bool {
	if !s.SC.filterChain.Fire(AfterConnected, s.context) {
		return false
	}
	s.SC.middleChain.Fire(AfterConnected, s.context)
	return true
}

func (s *Session) fireAfterClosed() bool {
	if !s.SC.filterChain.Fire(AfterClosed, s.context) {
		return false
	}
	s.SC.middleChain.Fire(AfterClosed, s.context)
	return true
}

func (s *Session) fireBeforeReceived() bool {
	if !s.SC.filterChain.Fire(BeforeReceived, s.context) {
		return false
	}
	s.SC.middleChain.Fire(BeforeReceived, s.context)
	return true
}

func (s *Session) fireAfterReceived() bool {
	if !s.SC.filterChain.Fire(AfterReceived, s.context) {
		return false
	}
	s.SC.middleChain.Fire(AfterReceived, s.context)
	return true
}

func (s *Session) fireBeforeSend() bool {
	if !s.SC.filterChain.Fire(BeforeSend, s.context) {
		return false
	}
	s.SC.middleChain.Fire(BeforeSend, s.context)
	return true
}

func (s *Session) fireAfterSend() bool {
	if !s.SC.filterChain.Fire(AfterSend, s.context) {
		return false
	}
	s.SC.middleChain.Fire(AfterSend, s.context)
	return true
}

func (s *Session) fireErrorMsgID() bool {
	if !s.SC.filterChain.Fire(ErrorMsgID, s.context) {
		return false
	}
	s.SC.middleChain.Fire(ErrorMsgID, s.context)
	return true
}

func (s *Session) fireSendMsgAfterSend(pack *sendPack) {
	if len(s.SC.filterChain.functions[AfterSend]) > 0 ||
		len(s.SC.middleChain.functions[AfterSend]) > 0 {
		msg := reflect.New(pack.msgType.Elem()).Interface()
		msgID, err := gMsgParser.UnmarshalUnregister(pack.data, msg, int(Config.LenMsgLen))
		putBuffer(bytes.NewBuffer(pack.data))
		if err != nil {
			log.Errorf("SendMsg UnmarshalUnregister error: %v", err)
			return
		}
		module.Obj.SendFunc(func(o *base.Object) {
			// update context
			s.context.MsgID = msgID
			s.context.Msg = msg
			s.fireAfterSend()
		})
	} else {
		putBuffer(bytes.NewBuffer(pack.data))
	}
}

func (s *Session) sendMsg() {
	s.agent.SendMsg()
}

func (s *Session) readMsg() {
	s.agent.ReadMsg()
}

func (s *Session) do() {
	for i := 0; i < s.SC.MaxRecv; i++ {
		select {
		case v := <-s.recv:
			msgID, msg, err := gMsgParser.Unmarshal(v, int(Config.LenMsgLen))
			if err != nil {
				var e *Error
				if errors.As(err, &e) && e.IsType(ErrorTypeMsgID) {
					// update context
					s.context.MsgID = msgID
					s.context.Packet = v[Config.LenMsgLen:]
					s.fireErrorMsgID()
					s.context.Packet = nil
					//todo v是否要回收再利用
				} else {
					log.Errorf("message unmarshal error: %v", err)
					putBuffer(bytes.NewBuffer(v))
				}
				continue
			}
			putBuffer(bytes.NewBuffer(v))

			// update context
			s.context.MsgID = msgID
			s.context.Msg = msg
			if !s.fireBeforeReceived() {
				continue
			}
			h := GetHandler(s.context.MsgID)
			if h != nil {
				h.Process(s.context)
				s.fireAfterReceived()
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
		defer func() { recover() }()
		close(s.closeSign)
	}

	err := s.agent.Close()

	select {
	case s.send <- nil:
	default:
	}
	return err
}

func (s *Session) String() string {
	return fmt.Sprintf("%v, ID:%v", s.SC, s.ID)
}
