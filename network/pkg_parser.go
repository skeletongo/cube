package network

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
)

// 应用层数据包解析器
//
// 封包结构
// --------------
// | len | data |
// --------------
// data 应用层数据包
// len 数据包长度（本身占用1或2或4个字节存储）

var gPkgParser = NewPkgParser()

// PkgParser 数据包解析器
type PkgParser struct {
	lenMsgLen uint32           // data数据的字节个数用几个字节存储
	minMsgLen uint32           // data数据最少字节个数
	maxMsgLen uint32           // data数据最长字节个数
	endian    binary.ByteOrder // 字节序
}

func NewPkgParser() *PkgParser {
	return &PkgParser{
		lenMsgLen: 2,
		minMsgLen: 1,
		maxMsgLen: 4096,
		endian:    binary.LittleEndian,
	}
}

// SetMsgLen 修改默认配置
// lenMsgLen 记录消息字节数占用的字节数
// minMsgLen 消息最少字节数量
// maxMsgLen 消息最大字节数量
// 返回校验后的配置
func (p *PkgParser) SetMsgLen(lenMsgLen uint32, minMsgLen uint32, maxMsgLen uint32) (uint32, uint32, uint32) {
	if lenMsgLen == 1 || lenMsgLen == 2 || lenMsgLen == 4 {
		p.lenMsgLen = lenMsgLen
	}
	if minMsgLen != 0 {
		p.minMsgLen = minMsgLen
	}
	if maxMsgLen != 0 {
		p.maxMsgLen = maxMsgLen
	}

	var max uint32
	switch p.lenMsgLen {
	case 1:
		max = math.MaxUint8
	case 2:
		max = math.MaxUint16
	case 4:
		max = math.MaxUint32
	}
	if p.minMsgLen > max {
		p.minMsgLen = max
	}
	if p.maxMsgLen > max {
		p.maxMsgLen = max
	}
	return p.lenMsgLen, p.minMsgLen, p.maxMsgLen
}

// SetByteOrder 修改字节序，默认小端序
func (p *PkgParser) SetByteOrder(order binary.ByteOrder) {
	p.endian = order
}

// Encode 应用层数据包编码
func (p *PkgParser) Encode(b []byte) (data []byte, err error) {
	var msgLen = uint32(len(b)) - p.lenMsgLen
	if msgLen > p.maxMsgLen {
		return nil, errors.New("message too long")
	}
	if msgLen < p.minMsgLen {
		return nil, errors.New("message too short")
	}
	switch p.lenMsgLen {
	case 1:
		b[0] = byte(msgLen)
	case 2:
		p.endian.PutUint16(b, uint16(msgLen))
	case 4:
		p.endian.PutUint32(b, msgLen)
	}
	return b, err
}

// EncodeByWriter 应用层数据包编码并发送
func (p *PkgParser) EncodeByWriter(w io.Writer, b []byte) (err error) {
	var data []byte
	data, err = p.Encode(b)
	if err != nil {
		return
	}
	_, err = w.Write(data)
	return
}

// Decode 协议层数据包解码成应用层数据包
func (p *PkgParser) Decode(b []byte) (data []byte, err error) {
	if len(b) < int(p.lenMsgLen) {
		return nil, errors.New("lenMsgLen too short")
	}

	var msgLen uint32
	switch p.lenMsgLen {
	case 1:
		msgLen = uint32(b[0])
	case 2:
		msgLen = uint32(p.endian.Uint16(b))
	case 4:
		msgLen = p.endian.Uint32(b)
	}

	if msgLen < p.minMsgLen {
		return nil, errors.New("message too short")
	}
	if msgLen > p.maxMsgLen {
		return nil, errors.New("message too long")
	}
	return b, err
}

// DecodeByReader 读取协议层数据包并解码成应用层数据包
func (p *PkgParser) DecodeByReader(r io.Reader) (b []byte, err error) {
	bs := getBytesN(int(p.lenMsgLen))
	defer func() {
		if err != nil {
			putBuffer(bytes.NewBuffer(bs))
		}
	}()

	if _, err = io.ReadFull(r, bs); err != nil {
		return nil, err
	}

	var msgLen uint32
	switch p.lenMsgLen {
	case 1:
		msgLen = uint32(bs[0])
	case 2:
		msgLen = uint32(p.endian.Uint16(bs))
	case 4:
		msgLen = p.endian.Uint32(bs)
	}

	if msgLen > p.maxMsgLen {
		return nil, errors.New("message too long")
	}
	if msgLen < p.minMsgLen {
		return nil, errors.New("message too short")
	}

	buf := bytes.NewBuffer(bs)
	buf.Grow(int(p.lenMsgLen + msgLen))
	bs = buf.Bytes()[:(int(p.lenMsgLen + msgLen))]
	_, err = io.ReadFull(r, bs[p.lenMsgLen:])
	return bs, err
}
