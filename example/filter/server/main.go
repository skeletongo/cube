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
	logrus.SetLevel(logrus.TraceLevel)

	network.SetHandlerFunc(1, &Ping{}, func(c *network.Context) {
		logrus.Info("ping:", c.Msg.(*Ping).Data)
		c.Send(2, &Pong{Data: "pong"})
	})
	cube.Run("config.json")
}
