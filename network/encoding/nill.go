package encoding

import "errors"

type Nil struct {
}

func (n *Nil) Unmarshal(buf []byte, data interface{}) error {
	return errors.New("nil EncodingType")
}

func (n *Nil) Marshal(data interface{}) ([]byte, error) {
	return nil, errors.New("nil EncodingType")
}
