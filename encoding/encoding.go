package encoding

import (
	"encoding/binary"

	"github.com/golang/protobuf/proto"
)

var gEncoding = NewEncoding()

// EncodeType 编码器类型
type EncodeType int

const (
	TypeNil EncodeType = iota
	TypeGPB
	TypeBinary
	TypeGob
	TypeJson
	TypeMax
)

// EncDecoder 应用层数据序列化方式
type EncDecoder interface {
	// Unmarshal 反序列化
	Unmarshal(buf []byte, data interface{}) error
	// Marshal 序列化
	Marshal(data interface{}) ([]byte, error)
}

// Encoding 编码管理器
type Encoding struct {
	DefaultEncodeType EncodeType
	encodingMap       [TypeMax]EncDecoder
}

// NewEncoding 创建编码管理器
func NewEncoding() *Encoding {
	return &Encoding{
		DefaultEncodeType: TypeGob,
		encodingMap: [TypeMax]EncDecoder{
			TypeNil:    new(Nil),
			TypeGPB:    new(GPB),
			TypeBinary: new(Binary),
			TypeGob:    new(Gob),
			TypeJson:   new(JSON),
		},
	}
}

// SetByteOrder 修改字节序，默认小端序
func (e *Encoding) SetByteOrder(order binary.ByteOrder) {
	e.encodingMap[TypeBinary] = &Binary{order}
}

// GetEncoding 获取指定的编码器
// 返回编码器和是否存在
func (e *Encoding) GetEncoding(n EncodeType) (EncDecoder, bool) {
	if n < 0 || n >= TypeMax {
		return nil, false
	}
	return e.encodingMap[n], true
}

// TypeTest 根据消息类型判断使用什么编码方式
func (e *Encoding) TypeTest(msg interface{}) EncodeType {
	switch msg.(type) {
	case proto.Message:
		return TypeGPB
	case []byte:
		return TypeBinary
	default:
		return e.DefaultEncodeType
	}
}

// SetByteOrder 修改字节序，默认小端序
func SetByteOrder(order binary.ByteOrder) {
	gEncoding.SetByteOrder(order)
}

// SetDefaultEncodeType 修改默认编码类型，默认 encoding/gob 编码
func SetDefaultEncodeType(n EncodeType) {
	gEncoding.DefaultEncodeType = n
}

// TypeTest 根据消息类型判断使用什么编码方式
func TypeTest(msg interface{}) EncodeType {
	return gEncoding.TypeTest(msg)
}

// GetEncoding 获取指定的编码器
// 返回编码器和是否存在
func GetEncoding(n EncodeType) (EncDecoder, bool) {
	return gEncoding.GetEncoding(n)
}
