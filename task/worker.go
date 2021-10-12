package task

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/stathat/consistent"

	"github.com/skeletongo/cube/base"
)

var gMaster *Master

// Worker 协程节点
type Worker struct {
	*base.Object
}

// Master 协程管理器
type Master struct {
	// 预创建协程的序号
	i int
	c *consistent.Consistent
	// 所有协程节点
	workers map[string]*Worker
	closing bool
}

func NewMaster(n int) *Master {
	m := &Master{
		c:       consistent.New(),
		workers: make(map[string]*Worker),
	}

	for i := 0; i < n; i++ {
		m.addWorker()
	}
	return m
}

func (m *Master) addWorker() *Worker {
	name := fmt.Sprintf("worker_%d", m.i)
	m.c.Add(name)
	return m.addWorkerByName(name)
}

func (m *Master) addWorkerByName(name string) *Worker {
	w := new(Worker)
	w.Object = base.NewObject(name, Config.Worker.Options, nil)
	w.Run()
	w.Data = w
	m.workers[w.Name] = w
	m.i++
	return w
}

func (m *Master) getWorker(name string) *Worker {
	workName, err := m.c.Get(name)
	if err != nil {
		return nil
	}
	return m.getWorkerByName(workName)
}

func (m *Master) getWorkerByName(name string) *Worker {
	if w, ok := m.workers[name]; ok {
		return w
	}
	return nil
}

// Close 关闭协程管理器
func Close() {
	if gMaster == nil || Obj == nil {
		return
	}
	Obj.SendFunc(func(o *base.Object) {
		if gMaster.closing {
			return
		}
		gMaster.closing = true
		for _, v := range gMaster.workers {
			v.Close()
		}
	})
	for _, v := range gMaster.workers {
		<-v.Closed
	}
	log.Info("task closed")
}
