package network

import (
	"bytes"
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

var gMsgParser = NewMsgParser()

// MsgParser 消息序列化和反序列化
type MsgParser struct {
	endian binary.ByteOrder
}

func NewMsgParser() *MsgParser {
	return &MsgParser{
		endian: binary.LittleEndian,
	}
}

func (m *MsgParser) SetByteOrder(order binary.ByteOrder) {
	m.endian = order
}

// Marshal 消息序列化
// msgID 消息号
// msg 消息数据
func (m *MsgParser) Marshal(msgID uint16, msg interface{}) ([]byte, error) {
	et := encoding.TypeTest(msg)
	p, _ := encoding.GetEncoding(et)
	data, err := p.Marshal(msg)
	if err != nil {
		return nil, err
	}

	bs := getBytesN(int(Config.LenMsgLen) + 4 + len(data))

	m.endian.PutUint16(bs[Config.LenMsgLen:], uint16(et))
	m.endian.PutUint16(bs[Config.LenMsgLen+2:], msgID)
	copy(bs[Config.LenMsgLen+4:], data)
	return bs, err
}

func (m *MsgParser) unmarshal(data []byte) (id, et uint16) {
	et = m.endian.Uint16(data)
	id = m.endian.Uint16(data[2:])
	return
}

// Unmarshal 消息反序列化
// data 序列化数据
// 返回消息号和解析后的消息数据
func (m *MsgParser) Unmarshal(data []byte) (msgID uint16, msg interface{}, err error) {
	defer putBuffer(bytes.NewBuffer(data))

	var et uint16
	msgID, et = m.unmarshal(data[Config.LenMsgLen:])
	msg = CreateMessage(msgID)
	if msg == nil {
		return 0, nil, NewError(errors.New("msgID unregister"), ErrorTypeMsgID, msgID)
	}
	p, has := encoding.GetEncoding(int(et))
	if !has {
		return 0, nil, NewError(errors.New("encoder error"), ErrorTypeEncoder, msgID)
	}
	return msgID, msg, p.Unmarshal(data[Config.LenMsgLen+4:], msg)
}

func (m *MsgParser) MarshalUnregister(msg interface{}) (data []byte, err error) {
	return m.Marshal(0, msg)
}

func (m *MsgParser) UnmarshalUnregister(data []byte, msg interface{}) (msgID uint16, err error) {
	defer putBuffer(bytes.NewBuffer(data))

	var et uint16
	msgID, et = m.unmarshal(data[Config.LenMsgLen:])
	p, has := encoding.GetEncoding(int(et))
	if !has {
		return msgID, NewError(errors.New("encoder error"), ErrorTypeEncoder, 0)
	}
	return msgID, p.Unmarshal(data[Config.LenMsgLen+4:], msg)
}
