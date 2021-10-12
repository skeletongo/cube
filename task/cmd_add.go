package task

import (
	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/base"
)

// sendToExecutor 给预创建的协程节点发送待执行的任务
func sendToExecutor(t *Task, name string) {
	if t == nil || Obj == nil {
		return
	}
	Obj.SendFunc(func(o *base.Object) {
		if gMaster.closing {
			log.Warning("task closed")
			return
		}
		w := gMaster.getWorker(name)
		if w == nil {
			log.Errorf("cannot find worker, name %s", name)
			return
		}

		sendCall(w.Object, t)
	})
}

// sendToFixExecutor 给指定的一个协程节点发送待执行的任务,如果协程节点找不到就新建一个
func sendToFixExecutor(t *Task, name string) {
	if t == nil || Obj == nil {
		return
	}
	Obj.SendFunc(func(o *base.Object) {
		if gMaster.closing {
			log.Warning("task closed")
			return
		}
		w := gMaster.getWorkerByName(name)
		if w == nil {
			// 创建新的协程节点
			w = gMaster.addWorkerByName(name)
		}

		sendCall(w.Object, t)
	})
}
