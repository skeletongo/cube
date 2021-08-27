package timer

import (
	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/base"
)

// SendTimer 执行延时方法
// o 执行节点
// t 延时方法
func SendTimer(o *base.Object, t *Timer) {
	if t == nil {
		log.Warningln("Timer is nil")
		return
	}
	if o == nil {
		log.Warnln("Timer error: object is nil")
		return
	}
	o.Send(base.CommandWrapper(func(o *base.Object) error {
		t.a.OnTimer(t.h, t.data)
		return nil
	}))
}
