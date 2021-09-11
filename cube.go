package cube

import (
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/module"
	"github.com/skeletongo/cube/network"
	"github.com/skeletongo/cube/pkg"
	"github.com/skeletongo/cube/task"
	"github.com/skeletongo/cube/timer"
)

func Run(config string) {
	log.Infof("Cube %v starting up", Version)

	// 需要启用的功能模块
	pkg.RegisterPackage(module.Config)
	pkg.RegisterPackage(task.Config)
	pkg.RegisterPackage(network.Config)

	// 读取配置文件，模块初始化
	pkg.Load(config)
	defer func() {
		pkg.Close()
		log.Infoln("Cube closed")
	}()

	timer.SetObject(module.Obj)
	task.SetObject(module.Obj)
	module.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Infof("Cube closing down (signal: %v)", sig)

	module.Close()
	task.Close()
	timer.StopAll()

	task.Obj.Close()
	<-task.Obj.Closed

	module.Obj.Close()
	<-module.Obj.Closed
}
