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

	network.SetHandlerFunc(1, &Ping{}, func(s *network.Session, msgID uint16, msg interface{}) error {
		logrus.Info("ping:", msg.(*Ping).Data)
		s.Send(2, &Pong{Data: "pong"})
		return nil
	})
	cube.Run("config.json")
}
