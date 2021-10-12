package main

import (
	"github.com/sirupsen/logrus"

	"github.com/skeletongo/cube"
	"github.com/skeletongo/cube/network"
)

type Ping struct {
	Data string
}

type Pong struct {
	Data string
}

func main() {
	network.AddMiddle(func() network.Middle {
		return &network.MiddleFunc{
			ErrorMsgID: func(c *network.Context) {
				logrus.Infof("--> msgID:%v Packet:%v", c.MsgID, string(c.Packet))
				msg := new(Ping)
				id, err := network.UnmarshalUnregister(c.Packet, msg)
				if err != nil {
					logrus.Error(err)
				}
				c.Send(2, &Pong{Data: "pong"})
				logrus.Infof("--> msgID:%v ping: %v", id, msg.Data)
				return
			},
		}
	})

	cube.Run("config.json")
}
