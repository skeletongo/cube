package network

import (
	"encoding/binary"
	"errors"

	"github.com/skeletongo/cube/encoding"
)

// 应用层消息解析器
//
// 应用层消息序列化结构
// -----------------------
// |EncodeType|MsgID|Data|
// -----------------------
// EncodeType 编解码类型
// MsgID 消息号
// Data 消息数据

// MsgParser 消息序列化和反序列化
type MsgParser struct {
	endian binary.ByteOrder
}

func NewMsgParser() *MsgParser {
	return &MsgParser{
		endian: binary.LittleEndian,
	}
}

// SetByteOrder 修改字节序，默认小端序
func (m *MsgParser) SetByteOrder(order binary.ByteOrder) {
	m.endian = order
}

// Marshal 消息序列化
// msgID 消息号
// msg 消息数据
// n 返回得数据切片前面填充几个空字节
func (m *MsgParser) Marshal(msgID uint16, msg interface{}, n int) ([]byte, error) {
	et := encoding.TypeTest(msg)
	p, _ := encoding.GetEncoding(et)
	data, err := p.Marshal(msg)
	if err != nil {
		return nil, err
	}

	bs := getBytesN(n + 4 + len(data))

	m.endian.PutUint16(bs[n:], uint16(et))
	m.endian.PutUint16(bs[n+2:], msgID)
	copy(bs[n+4:], data)
	return bs, err
}

func (m *MsgParser) unmarshal(data []byte) (id uint16, et encoding.EncodeType) {
	et = encoding.EncodeType(m.endian.Uint16(data))
	id = m.endian.Uint16(data[2:])
	return
}

// Unmarshal 消息解析
// data 序列化数据
// n 解析时跳过开头的几个字节
// 返回消息号和消息结构体的指针
func (m *MsgParser) Unmarshal(data []byte, n int) (msgID uint16, msg interface{}, err error) {
	var et encoding.EncodeType
	msgID, et = m.unmarshal(data[n:])
	msg = CreateMessage(msgID)
	if msg == nil {
		return msgID, nil, NewError(errors.New("msgID unregister"), ErrorTypeMsgID, msgID)
	}
	p, has := encoding.GetEncoding(et)
	if !has {
		return msgID, nil, NewError(errors.New("encoder error"), ErrorTypeEncoder, msgID)
	}
	return msgID, msg, p.Unmarshal(data[n+4:], msg)
}

// UnmarshalUnregister 未注册的消息解析
// data 序列化数据
// msg 消息结构体的指针
// n 解析时跳过开头的几个字节
// 返回消息号
func (m *MsgParser) UnmarshalUnregister(data []byte, msg interface{}, n int) (msgID uint16, err error) {
	var et encoding.EncodeType
	msgID, et = m.unmarshal(data[n:])
	p, has := encoding.GetEncoding(et)
	if !has {
		return msgID, NewError(errors.New("encoder error"), ErrorTypeEncoder, 0)
	}
	return msgID, p.Unmarshal(data[n+4:], msg)
}

var gMsgParser = NewMsgParser()

// Marshal 消息序列化
// msgID 消息号
// msg 消息数据
func Marshal(msgID uint16, msg interface{}) ([]byte, error) {
	return gMsgParser.Marshal(msgID, msg, 0)
}

// Unmarshal 消息解析
// data 序列化数据
// 返回消息号和消息结构体的指针
func Unmarshal(data []byte) (msgID uint16, msg interface{}, err error) {
	return gMsgParser.Unmarshal(data, 0)
}

// UnmarshalUnregister 未注册的消息解析
// data 序列化数据
// msg 消息结构体的指针
// 返回消息号
func UnmarshalUnregister(data []byte, msg interface{}) (msgID uint16, err error) {
	return gMsgParser.UnmarshalUnregister(data, msg, 0)
}
