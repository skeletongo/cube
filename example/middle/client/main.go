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
	logrus.SetLevel(logrus.TraceLevel)

	// 方式一:配置文件控制启用及调用顺序
	network.RegisterMiddle("test_middle1", func() network.Middle {
		return &network.MiddleFunc{
			AfterConnected: func(c *network.Context) {
				c.Set("n", 0)
				logrus.Info("--> AfterConnected")
				c.Send(1, &Ping{Data: "ping"})
			},
		}
	})
	network.RegisterMiddle("test_middle2", func() network.Middle {
		return new(TestMiddle)
	})

	// 方式二:代码中控制启用及调用顺
	network.AddMiddle(func() network.Middle {
		return &network.MiddleFunc{
			AfterConnected: func(c *network.Context) {
				c.Set("n", 0)
				logrus.Info("--> AfterConnected")
				c.Send(1, &Ping{Data: "ping"})
			},
		}
	})
	network.AddMiddle(func() network.Middle {
		return new(TestMiddle)
	})

	network.SetHandlerFunc(2, &Pong{}, func(ctx *network.Context) {
		logrus.Info("pong:", ctx.Msg.(*Pong).Data)
		timer.AfterTimer(time.Second*3, func() {
			ctx.Send(1, &Ping{Data: "ping"})
		})
	})
	cube.Run()
}

type TestMiddle struct {
}

func (m *TestMiddle) Get(op network.Opportunity) func(c *network.Context) {
	switch op {
	case network.AfterClosed:
		return func(c *network.Context) {
			logrus.Infof("--> AfterClosed SessionInfo:%v", c.Session)
		}

	case network.BeforeReceived:
		return func(c *network.Context) {
			c.Set("n", c.GetInt("n")+1)
			logrus.Infof("--> BeforeReceived MsgID:%v Msg:%v N:%v", c.MsgID, c.Msg, c.GetInt("n"))
		}

	case network.AfterReceived:
		return func(c *network.Context) {
			logrus.Infof("--> AfterReceived MsgID:%v Msg:%v N:%v", c.MsgID, c.Msg, c.GetInt("n"))
		}

	case network.BeforeSend:
		return func(c *network.Context) {
			logrus.Infof("--> BeforeSend MsgID:%v Msg:%v N:%v", c.MsgID, c.Msg, c.GetInt("n"))
		}

	case network.AfterSend:
		return func(c *network.Context) {
			logrus.Infof("--> AfterSend MsgID:%v Msg:%v N:%v", c.MsgID, c.Msg, c.GetInt("n"))
		}
	}
	return nil
}
