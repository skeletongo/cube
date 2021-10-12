package network

import (
	"reflect"

	log "github.com/sirupsen/logrus"
)

// Handler 消息处理接口
type Handler interface {
	Process(c *Context)
}

type HandlerWrapper func(c *Context)

func (hw HandlerWrapper) Process(c *Context) {
	hw(c)
}

type MsgInfo struct {
	msgType    reflect.Type
	msgHandler Handler
}

// MsgHandler 消息注册表
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
// 返回消息结构体的指针
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
func (m *MsgHandler) SetHandlerFunc(msgID uint16, msg interface{}, handlerFunc func(c *Context)) {
	m.SetHandler(msgID, msg, HandlerWrapper(handlerFunc))
}

var gMsgHandler = NewMsgHandler()

// CreateMessage 根据消息号创建对应的消息实例
// msgID 消息号
// 返回消息结构体的指针
func CreateMessage(msgID uint16) interface{} {
	return gMsgHandler.CreateMessage(msgID)
}

// GetHandler 根据消息号获取消息处理方法
// msgID 消息号
// handler 消息处理方法
func GetHandler(msgID uint16) Handler {
	return gMsgHandler.GetHandler(msgID)
}

// SetHandler 设置消息处理方法
// msgID 消息号
// msg 消息结构体指针
// handler 消息处理方法
func SetHandler(msgID uint16, msg interface{}, handler Handler) {
	gMsgHandler.SetHandler(msgID, msg, handler)
}

// SetHandlerFunc 设置消息处理方法
// msgID 消息号
// msg 消息结构体指针
// handlerFunc 消息处理方法
func SetHandlerFunc(msgID uint16, msg interface{}, handlerFunc func(c *Context)) {
	gMsgHandler.SetHandlerFunc(msgID, msg, handlerFunc)
}
