package task

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/base"
)

var ErrCannotFindWorker = errors.New("Cannot find Worker ")

// sendToExecutor 给预创建的协程节点发送待执行的任务
func sendToExecutor(t *Task, name string) {
	if t == nil || Obj == nil {
		return
	}
	Obj.Send(base.CommandWrapper(func(o *base.Object) error {
		if gMaster.closing {
			log.WithField("name", t.Name).Warning("Task closed")
			return nil
		}
		w := gMaster.getWorker(name)
		if w == nil {
			return ErrCannotFindWorker
		}

		sendCall(w.Object, t)
		return nil
	}))
}

// sendToFixExecutor 给指定的一个协程节点发送待执行的任务,如果协程节点找不到就新建一个
func sendToFixExecutor(t *Task, name string) {
	if t == nil || Obj == nil {
		return
	}
	Obj.Send(base.CommandWrapper(func(o *base.Object) error {
		if gMaster.closing {
			log.WithField("name", t.Name).Warning("Task closed")
			return nil
		}
		w := gMaster.getWorkerByName(name)
		if w == nil {
			// 创建新的协程节点
			w = gMaster.addWorkerByName(name)
		}

		sendCall(w.Object, t)
		return nil
	}))
}
