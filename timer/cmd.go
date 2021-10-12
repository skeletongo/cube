package timer

import (
	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/base"
)

// SendTimer 执行延时方法
// o 执行节点
// t 延时方法
func SendTimer(o *base.Object, f func()) {
	if o == nil {
		log.Warning("timer error: object is nil")
		return
	}
	o.SendFunc(func(o *base.Object) {
		f()
	})
}
