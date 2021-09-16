package main

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skeletongo/cube"
	"github.com/skeletongo/cube/module"
	"github.com/skeletongo/cube/utils"
)

type myModule struct {
}

func (m *myModule) Name() string {
	return "testModule"
}

func (m *myModule) Init() {
	panic("my module panic")
}

func (m *myModule) Update() {
	logrus.Error("eeeeeeeeeeeeeeeee")
	panic("my module panic")
}

func (m *myModule) Close() {
	defer module.Closed(m)
	panic("my module panic")
}

func main() {
	logrus.AddHook(utils.NewFileLineHook(logrus.ErrorLevel))
	module.Register(new(myModule), time.Second*5, 0)
	cube.Run("config.json")
}
