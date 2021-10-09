package encoding

import (
	"encoding/binary"

	"github.com/golang/protobuf/proto"
)

var gEncoding = NewEncoding()

const (
	TypeNil = iota
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

type Encoding struct {
	DefaultEncodeType int
	encodingMap       [TypeMax]EncDecoder
}

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

func (e *Encoding) SetByteOrder(order binary.ByteOrder) {
	e.encodingMap[TypeBinary] = &Binary{order}
}

func (e *Encoding) GetEncoding(n int) (EncDecoder, bool) {
	if n < 0 || n >= TypeMax {
		return nil, false
	}
	return e.encodingMap[n], true
}

func (e *Encoding) TypeTest(msg interface{}) int {
	switch msg.(type) {
	case proto.Message:
		return TypeGPB
	case []byte:
		return TypeBinary
	default:
		return e.DefaultEncodeType
	}
}

func SetByteOrder(order binary.ByteOrder) {
	gEncoding.SetByteOrder(order)
}

func SetDefaultEncodeType(n int) {
	gEncoding.DefaultEncodeType = n
}

func TypeTest(msg interface{}) int {
	return gEncoding.TypeTest(msg)
}

func GetEncoding(n int) (EncDecoder, bool) {
	return gEncoding.GetEncoding(n)
}
