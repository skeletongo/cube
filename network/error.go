package network

import "fmt"

// ErrParserPacket 消息类型不支持或解析失败
type ErrParserPacket struct {
	EncodeType uint16 // 消息类型
	MsgID      uint16 // 消息号
	Err        error  // 错误信息
}

func (e *ErrParserPacket) Error() string {
	return fmt.Sprintf("cannot parse proto type:%v msgID:%v error:%v", e.EncodeType, e.MsgID, e.Err)
}

func NewErrParsePacket(et, msgID uint16, err error) *ErrParserPacket {
	return &ErrParserPacket{EncodeType: et, MsgID: msgID, Err: err}
}
