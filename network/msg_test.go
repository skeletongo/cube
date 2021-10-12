package network_test

import (
	"testing"

	"github.com/skeletongo/cube/network"
)

var gMsgParser = network.NewMsgParser()

type D struct {
	Name string
	Age  int
}

func TestMarshal(t *testing.T) {
	network.SetHandlerFunc(1, new(D), func(c *network.Context) {
	})

	data, err := gMsgParser.Marshal(1, &D{
		Name: "Tom",
		Age:  20,
	}, 2)
	if err != nil {
		t.Error(err)
		return
	}

	id, msg, err := gMsgParser.Unmarshal(data, 2)
	t.Logf("msgID:%v Msg:%v Err:%v\n", id, *msg.(*D), err)
}

func TestUnmarshalUnregister(t *testing.T) {
	data, err := gMsgParser.Marshal(1, &D{
		Name: "Tom",
		Age:  20,
	}, 2)
	if err != nil {
		t.Error(err)
		return
	}

	msg := new(D)

	id, err := gMsgParser.UnmarshalUnregister(data, msg, 2)
	t.Logf("msgID:%v Msg:%v Err:%v\n", id, *msg, err)
}
