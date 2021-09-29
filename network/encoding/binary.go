package encoding

import (
	"bytes"
	"encoding/binary"
)

// Binary 二进制
type Binary struct {
}

func (b *Binary) Unmarshal(buf []byte, data interface{}) error {
	return binary.Read(bytes.NewReader(buf), defaultEndian, data)
}

func (b *Binary) Marshal(data interface{}) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	err := binary.Write(buffer, defaultEndian, data)
	return buffer.Bytes(), err
}
