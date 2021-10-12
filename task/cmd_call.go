package task

import (
	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/base"
)

// sendCall 给协程节点发送要执行的任务
// o 需要执行 Task.callFunc 的节点
func sendCall(o *base.Object, t *Task) {
	if t == nil {
		log.Error("task is nil")
		return
	}
	if o == nil {
		log.Error("task run CallFunc error: object is nil")
		return
	}
	o.SendFunc(func(o *base.Object) {
		t.call()
	})
}
