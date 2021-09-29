package network

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
)

// 封包结构
// --------------
// | len | data |
// --------------
// data 业务数据
// len 业务数据字节长度（本身占用1或2或4个字节存储）

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

var gPkgParser = NewPkgParser()

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

func (p *PkgParser) SetByteOrder(order binary.ByteOrder) {
	p.endian = order
}

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

func (p *PkgParser) EncodeByIOWriter(w io.Writer, b []byte) (err error) {
	defer putBuffer(bytes.NewBuffer(b))

	var data []byte
	data, err = p.Encode(b)
	if err != nil {
		return
	}
	_, err = w.Write(data)
	return
}

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

func (p *PkgParser) DecodeByIOReader(r io.Reader) (b []byte, err error) {
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
