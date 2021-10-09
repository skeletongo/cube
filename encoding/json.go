package encoding

import "encoding/json"

type JSON struct {
}

func (J *JSON) Unmarshal(buf []byte, data interface{}) error {
	return json.Unmarshal(buf, data)
}

func (J *JSON) Marshal(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}
