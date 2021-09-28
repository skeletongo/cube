package network_test

import (
	"fmt"
	"testing"

	"github.com/skeletongo/cube/network"
)

var gMsgParser = network.NewMsgParser()

type D struct {
	Name string
	Age  int
}

func TestMarshal(t *testing.T) {
	network.SetHandlerFunc(1, new(D), func(s *network.Session, msgID uint16, msg interface{}) error {
		return nil
	})

	data, err := gMsgParser.Marshal(1, &D{
		Name: "Tom",
		Age:  20,
	})
	if err != nil {
		t.Error(err)
		return
	}

	id, msg, err := gMsgParser.Unmarshal(data)
	t.Logf("msgID:%v Msg:%v Err:%v\n", id, *msg.(*D), err)
}

func TestMarshalNoMsgID(t *testing.T) {
	data, err := gMsgParser.MarshalNoMsgID(&D{
		Name: "Tom",
		Age:  20,
	})
	if err != nil {
		t.Error(err)
		return
	}

	msg := new(D)
	err = gMsgParser.UnmarshalNoMsgID(data, msg)
	fmt.Printf("Msg:%v Err:%v\n", msg, err)
}
