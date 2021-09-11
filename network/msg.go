package network

import (
	"encoding/binary"
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/network/encoding"
)

/*
 消息序列化结构
 -----------------------
 |EncodeType|MsgID|Data|
 -----------------------
 EncodeType 编解码类型
 MsgID 消息号
 Data 消息数据
*/

// MsgParser 消息序列化和反序列化
type MsgParser struct {
	endian binary.ByteOrder
}

func NewMsgParser() *MsgParser {
	return &MsgParser{
		endian: binary.LittleEndian,
	}
}

var gMsgParser = NewMsgParser()

func (m *MsgParser) SetByteOrder(order binary.ByteOrder) {
	m.endian = order
}

// Marshal 消息序列化
// msgID 消息号
// msg 消息数据
func (m *MsgParser) Marshal(msgID uint16, msg interface{}) ([][]byte, error) {
	et := encoding.TypeTest(msg)
	data, err := encoding.Encoding[et].Marshal(msg)
	if err != nil {
		return nil, err
	}

	head := make([]byte, 4)
	m.endian.PutUint16(head, uint16(et))
	m.endian.PutUint16(head[2:], msgID)
	return [][]byte{head, data}, err
}

func (m *MsgParser) unmarshal(data []byte) (id, et uint16, err error) {
	et = m.endian.Uint16(data)
	id = m.endian.Uint16(data[2:])
	if et < encoding.TypeNil || et >= encoding.TypeMax {
		return id, et, NewErrParsePacket(et, id, fmt.Errorf("EncodeType:%d unregiste", et))
	}
	return
}

// Unmarshal 消息反序列化
// data 序列化数据
// 返回消息号和解析后的消息数据
func (m *MsgParser) Unmarshal(data []byte) (msgID uint16, msg interface{}, err error) {
	var et uint16
	msgID, et, err = m.unmarshal(data)
	msg = CreateMessage(msgID)
	if msg == nil {
		return 0, nil, NewErrParsePacket(et, msgID, fmt.Errorf("MsgID:%d unregiste", msgID))
	}
	return msgID, msg, encoding.Encoding[et].Unmarshal(data[4:], msg)
}

func (m *MsgParser) MarshalNoMsgID(msg interface{}) (data [][]byte, err error) {
	return m.Marshal(0, msg)
}

func (m *MsgParser) UnmarshalNoMsgID(data []byte, msg interface{}) error {
	_, et, err := m.unmarshal(data)
	if err != nil {
		return err
	}
	return encoding.Encoding[et].Unmarshal(data[4:], msg)
}

// ===============================
// 业务层消息处理方法的注册
// ===============================

// Handler 消息处理接口
type Handler interface {
	// Process 处理收到的消息
	// s 连接
	// msgID 消息号
	// msg 消息内容
	Process(s *Session, msgID uint16, msg interface{}) error
}

type handlerWrapper func(s *Session, msgID uint16, msg interface{}) error

func (hw handlerWrapper) Process(s *Session, msgID uint16, msg interface{}) error {
	return hw(s, msgID, msg)
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
func (m *MsgHandler) SetHandlerFunc(msgID uint16, msg interface{},
	handlerFunc func(s *Session, msgID uint16, msg interface{}) error) {
	m.SetHandler(msgID, msg, handlerWrapper(handlerFunc))
}

var gMsgHandler = NewMsgHandler()

func CreateMessage(msgID uint16) interface{} {
	return gMsgHandler.CreateMessage(msgID)
}

func GetHandler(msgID uint16) Handler {
	return gMsgHandler.GetHandler(msgID)
}

func SetHandler(msgID uint16, msg interface{}, handler Handler) {
	gMsgHandler.SetHandler(msgID, msg, handler)
}

func SetHandlerFunc(msgID uint16, msg interface{},
	handlerFunc func(s *Session, msgID uint16, msg interface{}) error) {
	gMsgHandler.SetHandlerFunc(msgID, msg, handlerFunc)
}
