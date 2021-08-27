package task

import (
	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/base"
)

// sendCallback 执行回调函数
func sendCallback(o *base.Object, t *Task) {
	if t == nil {
		log.Warningln("Task is nil")
		return
	}
	if o == nil {
		log.WithField("name", t.Name).Errorln("Task run CallbackFunc error: object is nil")
		return
	}
	o.Send(base.CommandWrapper(func(o *base.Object) error {
		t.callbackFunc(t.ret, t)
		return nil
	}))
}
