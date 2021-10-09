package network

import (
	"reflect"

	log "github.com/sirupsen/logrus"
)

// 应用层消息管理器

var gMsgHandler = NewMsgHandler()

// Handler 消息处理接口
type Handler interface {
	Process(c *Context)
}

type handlerWrapper func(c *Context) error

func (hw handlerWrapper) Process(c *Context) {
	hw(c)
}

type MsgInfo struct {
	msgType    reflect.Type
	msgHandler Handler
}

type MsgHandler struct {
	messages map[uint16]*MsgInfo
}

func NewMsgHandler() *MsgHandler {
	return &MsgHandler{
		messages: make(map[uint16]*MsgInfo),
	}
}

// CreateMessage 根据消息号创建对应的消息实例
// msgID 消息号
func (m *MsgHandler) CreateMessage(msgID uint16) interface{} {
	v, ok := m.messages[msgID]
	if !ok || v.msgType == nil {
		return nil
	}
	return reflect.New(v.msgType.Elem()).Interface()
}

// GetHandler 根据消息号获取消息处理方法
// msgID 消息号
// handler 消息处理方法
func (m *MsgHandler) GetHandler(msgID uint16) Handler {
	v, ok := m.messages[msgID]
	if !ok {
		return nil
	}
	return v.msgHandler
}

// SetHandler 设置消息处理方法
// msgID 消息号
// msg 消息结构体指针
// handler 消息处理方法
func (m *MsgHandler) SetHandler(msgID uint16, msg interface{}, handler Handler) {
	if _, ok := m.messages[msgID]; ok {
		log.WithField("msgID", msgID).Panicln("message already exist")
		return
	}

	msgType := reflect.TypeOf(msg)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		log.WithField("msgID", msgID).Panicln("message pointer required")
		return
	}

	if handler == nil {
		log.WithField("msgID", msgID).Panicln("message handler is nil")
		return
	}

	m.messages[msgID] = &MsgInfo{
		msgType:    msgType,
		msgHandler: handler,
	}
}

// SetHandlerFunc 设置消息处理方法
// msgID 消息号
// msg 消息结构体指针
// handlerFunc 消息处理方法
func (m *MsgHandler) SetHandlerFunc(msgID uint16, msg interface{}, handlerFunc func(c *Context) error) {
	m.SetHandler(msgID, msg, handlerWrapper(handlerFunc))
}

func CreateMessage(msgID uint16) interface{} {
	return gMsgHandler.CreateMessage(msgID)
}

func GetHandler(msgID uint16) Handler {
	return gMsgHandler.GetHandler(msgID)
}

func SetHandler(msgID uint16, msg interface{}, handler Handler) {
	gMsgHandler.SetHandler(msgID, msg, handler)
}

func SetHandlerFunc(msgID uint16, msg interface{}, handlerFunc func(c *Context) error) {
	gMsgHandler.SetHandlerFunc(msgID, msg, handlerFunc)
}
