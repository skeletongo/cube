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
	network.RegisterFilter("test_filter1", func() network.Filter {
		return &network.FilterFunc{
			AfterConnected: func(c *network.Context) bool {
				c.Set("n", 0)
				logrus.Info("--> AfterConnected")
				c.Send(1, &Ping{Data: "ping"})
				return true
			},
		}
	})
	network.RegisterFilter("test_filter2", func() network.Filter {
		return new(TestFilter)
	})

	// 方式二:代码中控制启用及调用顺
	network.AddFilter(func() network.Filter {
		return &network.FilterFunc{
			AfterConnected: func(c *network.Context) bool {
				c.Set("n", 0)
				logrus.Info("--> AfterConnected")
				c.Send(1, &Ping{Data: "ping"})
				return true
			},
		}
	})
	network.AddFilter(func() network.Filter {
		return new(TestFilter)
	})

	network.SetHandlerFunc(2, &Pong{}, func(ctx *network.Context) error {
		logrus.Info("pong:", ctx.Msg.(*Pong).Data)
		timer.AfterTimer(time.Second*3, func() {
			ctx.Send(1, &Ping{Data: "ping"})
		})
		return nil
	})
	cube.Run("config.json")
}

type TestFilter struct {
}

func (m *TestFilter) Get(op network.Opportunity) func(c *network.Context) bool {
	switch op {
	case network.AfterClosed:
		return func(c *network.Context) bool {
			logrus.Infof("--> AfterClosed SessionInfo:%v", c.Session)
			return true
		}

	case network.BeforeReceived:
		return func(c *network.Context) bool {
			if c.GetInt("n") >= 5 {
				c.Close()
				return false
			}
			c.Set("n", c.GetInt("n")+1)
			logrus.Infof("--> BeforeReceived MsgID:%v Msg:%v N:%v", c.MsgID, c.Msg, c.GetInt("n"))
			return true
		}

	case network.AfterReceived:
		return func(c *network.Context) bool {
			logrus.Infof("--> AfterReceived MsgID:%v Msg:%v", c.MsgID, c.Msg)
			return true
		}

	case network.BeforeSend:
		return func(c *network.Context) bool {
			logrus.Infof("--> BeforeSend MsgID:%v Msg:%v", c.MsgID, c.Msg)
			return true
		}

	case network.AfterSend:
		return func(c *network.Context) bool {
			logrus.Infof("--> AfterSend MsgID:%v Msg:%v", c.MsgID, c.Msg)
			return true
		}
	}
	return nil
}
