package main

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/skeletongo/cube"
	"github.com/skeletongo/cube/network"
	"github.com/skeletongo/cube/timer"
)

type Ping struct {
	Data string
}

type Pong struct {
	Data string
}

func main() {
	logrus.SetLevel(logrus.InfoLevel)

	network.RegisterFilter("test_filter", func() network.Filter {
		return new(myFilter)
	})

	network.SetHandlerFunc(2, &Pong{}, func(ctx *network.Context) {
		logrus.Info("pong:", ctx.Msg.(*Pong).Data)
		timer.AfterTimer(time.Second*3, func() {
			ctx.Send(1, &Ping{Data: "ping"})
		})
	})
	cube.Run("config.json")
}

type myFilter struct {
}

func (m *myFilter) Get(op network.Opportunity) func(c *network.Context) bool {
	switch op {
	case network.AfterConnected:
		return func(c *network.Context) bool {
			logrus.Info("--> AfterConnected")
			c.Send(1, &Ping{Data: "ping"})
			return true
		}
	}
	return nil
}
