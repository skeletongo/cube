package encoding

import (
	"encoding/binary"

	"github.com/golang/protobuf/proto"
)

var defaultEndian binary.ByteOrder = binary.LittleEndian

// SetEndian 设置大小端序
func SetEndian(endian binary.ByteOrder) {
	defaultEndian = endian
}

var defaultEncodeType = TypeGob

func SetDefaultEncodeType(n int) {
	defaultEncodeType = n
}

// EncDecoder 应用层数据序列化方式
type EncDecoder interface {
	// Unmarshal 反序列化
	Unmarshal(buf []byte, data interface{}) error
	// Marshal 序列化
	Marshal(data interface{}) ([]byte, error)
}

const (
	TypeNil = iota
	TypeGPB
	TypeBinary
	TypeGob
	TypeJson
	TypeMax
)

var Encoding = [TypeMax]EncDecoder{
	TypeNil:    new(Nil),
	TypeGPB:    new(GPB),
	TypeBinary: new(Binary),
	TypeGob:    new(Gob),
	TypeJson:   new(JSON),
}

// TypeTest 消息编码类型判断
func TypeTest(msg interface{}) int {
	switch msg.(type) {
	case proto.Message:
		return TypeGPB
	case []byte:
		return TypeBinary
	default:
		return defaultEncodeType
	}
}
