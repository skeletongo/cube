package encoding

import (
	"errors"

	"google.golang.org/protobuf/proto"
)

var ErrorTypeNotFit = errors.New("msg not proto.Message type")

// GPB Google's protocol
type GPB struct {
}

func (p *GPB) Unmarshal(buf []byte, data interface{}) error {
	protoMsg, ok := data.(proto.Message)
	if !ok {
		return ErrorTypeNotFit
	}
	return proto.Unmarshal(buf, protoMsg)
}

func (p *GPB) Marshal(data interface{}) ([]byte, error) {
	protoMsg, ok := data.(proto.Message)
	if !ok {
		return nil, ErrorTypeNotFit
	}
	return proto.Marshal(protoMsg)
}
