package encoding

import (
	"bytes"
	"encoding/binary"
)

// Binary 二进制编码
type Binary struct {
	ByteOrder binary.ByteOrder
}

func (b *Binary) Unmarshal(buf []byte, data interface{}) error {
	return binary.Read(bytes.NewReader(buf), b.ByteOrder, data)
}

func (b *Binary) Marshal(data interface{}) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	err := binary.Write(buffer, b.ByteOrder, data)
	return buffer.Bytes(), err
}
