package encoding

import (
	"bytes"
	"encoding/gob"
)

// Gob encoding/gob
type Gob struct {
}

func (d *Gob) Unmarshal(buf []byte, data interface{}) error {
	return gob.NewDecoder(bytes.NewReader(buf)).Decode(data)
}

func (d *Gob) Marshal(data interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(data)
	return buf.Bytes(), err
}
