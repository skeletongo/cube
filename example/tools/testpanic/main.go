package main

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/skeletongo/cube"
	"github.com/skeletongo/cube/module"
	"github.com/skeletongo/cube/tools"
)

type myModule struct {
}

func (m *myModule) Name() string {
	return "testModule"
}

func (m *myModule) Init() {
	panic("my module panic")
}

func (m *myModule) AfterInit() {
	panic("my module panic")
}

func (m *myModule) Update() {
	logrus.Error("error line 25")
	logrus.Errorf("error line 26")
	logrus.Errorln("error line 27")
	panic("my module panic")
}

func (m *myModule) BeforeClose() {}

func (m *myModule) Close() {
	defer module.Release(m)
	panic("my module panic")
}

func main() {
	*logrus.StandardLogger() = *logrus.New()

	//h := &tools.FileLineHook{
	//	LogLevels: []logrus.Level{logrus.ErrorLevel},
	//	FieldName: "line",
	//	Skip:      8,
	//	Num:       2,
	//	Test:      false,
	//}

	logrus.AddHook(tools.NewFileLineHook(logrus.AllLevels...))

	cube.Register(module.Config)
	module.Register(new(myModule), time.Second*5, 0)
	cube.Load()
	module.Start()
	<-module.Obj.Closed
}
