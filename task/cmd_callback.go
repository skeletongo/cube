package task

import (
	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/base"
)

// sendCallback 执行回调函数
func sendCallback(o *base.Object, t *Task) {
	if t == nil {
		log.Error("task is nil")
		return
	}
	if o == nil {
		log.Error("task run CallbackFunc error: object is nil")
		return
	}
	o.SendFunc(func(o *base.Object) {
		t.callback()
	})
}
