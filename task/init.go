package task

import (
	"github.com/skeletongo/cube/base"
)

// Obj 协程管理节点，所有新协程的创建都是由这个节点完成
var Obj *base.Object

// Config 配置
var Config = new(Configuration)

type WorkerConfig struct {
	Options   *base.Options // 协程节点配置
	WorkerCnt int           // 预创建的协程数量
}

type Configuration struct {
	Options *base.Options // 协程管理节点配置
	Worker  *WorkerConfig // 协程节点配置
}

func (c *Configuration) Name() string {
	return "task"
}

func (c *Configuration) Init() error {
	Obj = base.NewObject("task", c.Options, nil)
	Obj.Run()

	if c.Worker.WorkerCnt <= 0 {
		c.Worker.WorkerCnt = 4
	}
	// 预创建协程节点，并连接到 Obj 节点，作为子节点
	gMaster = NewMaster(c.Worker.WorkerCnt)
	return nil
}

func (c *Configuration) Close() error {
	return nil
}
