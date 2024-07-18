package cube

import (
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/g"
	"github.com/skeletongo/cube/module"
	"github.com/skeletongo/cube/network"
	"github.com/skeletongo/cube/statsviz"
	"github.com/skeletongo/cube/timer"
)

func Run() {
	log.Infof("Cube %v starting up", Version)

	// 需要启用的功能模块
	Register(module.Config)
	Register(network.Config)
	Register(statsviz.Config)

	// 读取配置文件，模块初始化
	Load()
	defer func() {
		Close()
		log.Infoln("Cube closed")
	}()

	timer.SetObject(module.Obj)
	g.SetObject(module.Obj)
	module.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Infof("Cube closing down (signal: %v)", sig)

	module.Close()
	g.Close()
	timer.StopAll()

	module.Obj.Close()
	<-module.Obj.Closed
}
