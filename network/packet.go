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
	lenMsgLen int              // data数据的字节个数用几个字节存储
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

func (p *PkgParser) SetMsgLen(lenMsgLen int, minMsgLen uint32, maxMsgLen uint32) {
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
}

func (p *PkgParser) SetByteOrder(order binary.ByteOrder) {
	p.endian = order
}

func (p *PkgParser) Encode(w io.Writer, args [][]byte) (err error) {
	var msgLen uint32
	for i := 0; i < len(args); i++ {
		msgLen += uint32(len(args[i]))
	}

	if msgLen > p.maxMsgLen {
		return errors.New("message too long")
	}
	if msgLen < p.minMsgLen {
		return errors.New("message too short")
	}

	var head []byte
	switch p.lenMsgLen {
	case 1:
		head = []byte{byte(msgLen)}
	case 2:
		head = make([]byte, 2)
		p.endian.PutUint16(head, uint16(msgLen))
	case 4:
		head = make([]byte, 4)
		p.endian.PutUint32(head, msgLen)
	}

	readers := make([]io.Reader, 0, len(args)+1)
	readers = append(readers, bytes.NewReader(head))
	for i := 0; i < len(args); i++ {
		readers = append(readers, bytes.NewReader(args[i]))
	}
	_, err = io.Copy(w, io.MultiReader(readers...))
	return
}

func (p *PkgParser) Decode(r io.Reader) (data []byte, err error) {
	var b [4]byte
	bufMsgLen := b[:p.lenMsgLen]

	if _, err := io.ReadFull(r, bufMsgLen); err != nil {
		return nil, err
	}

	var msgLen uint32
	switch p.lenMsgLen {
	case 1:
		msgLen = uint32(bufMsgLen[0])
	case 2:
		msgLen = uint32(p.endian.Uint16(bufMsgLen))
	case 4:
		msgLen = p.endian.Uint32(bufMsgLen)
	}

	if msgLen > p.maxMsgLen {
		return nil, errors.New("message too long")
	}
	if msgLen < p.minMsgLen {
		return nil, errors.New("message too short")
	}

	data = make([]byte, msgLen)
	_, err = io.ReadFull(r, data)
	return
}
