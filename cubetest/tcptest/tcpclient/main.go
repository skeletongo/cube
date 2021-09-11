package main

import (
	"github.com/sirupsen/logrus"
	"github.com/skeletongo/cube"
	"github.com/skeletongo/cube/network"
	"github.com/skeletongo/cube/timer"
	"time"
)

type Ping struct {
	Data string
}

type Pong struct {
	Data string
}

func main() {
	logrus.SetLevel(logrus.TraceLevel)

	time.AfterFunc(time.Second, func() {
		if network.TestTCPClientSession == nil {
			logrus.Info("no send")
		} else {
			network.TestTCPClientSession.Send(1, &Ping{Data: "ping"})
		}
	})

	network.SetHandlerFunc(2, &Pong{}, func(s *network.Session, msgID uint16, msg interface{}) error {
		logrus.Info("pong:", msg.(*Pong).Data)
		timer.AfterTimer(time.Second*3, func() {
			s.Send(1, &Ping{Data: "ping"})
		})
		return nil
	})
	cube.Run("config.json")
}
