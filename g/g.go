package g

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/base"
	"github.com/skeletongo/cube/tools"
)

// gObject 回调方法默认执行节点
var gObject *base.Object

func SetObject(o *base.Object) {
	gObject = o
}

// num 启动的协程数量
var num int64

// Wait 等待所有协程处理完成
func Wait() {
	if atomic.LoadInt64(&num) == 0 {
		log.Info("goroutines closed")
		return
	}

	t := time.NewTicker(time.Second)
	for {
		select {
		case <-t.C:
			n := atomic.LoadInt64(&num)
			if n == 0 {
				log.Info("goroutines closed")
				return
			}
			log.Infof("goroutines closing, remaining %d", n)
		}
	}
}

type g struct {
	o *base.Object
}

// New 创建协程对象
// o 回调方法执行节点
func New(o ...*base.Object) *g {
	var obj *base.Object
	if len(o) > 0 {
		obj = o[0]
	}
	if obj == nil {
		obj = gObject
	}
	return &g{
		o: obj,
	}
}

// Go 启动一个协程
// callFunc 在协程中执行的方法
// callbackFunc 回调方法
func (g *g) Go(callFunc func(), callbackFunc ...func()) {
	var f func()
	if len(callbackFunc) > 0 {
		f = callbackFunc[0]
	}

	atomic.AddInt64(&num, 1)

	go func() {
		defer func() {
			if g.o == nil || f == nil {
				atomic.AddInt64(&num, -1)
			} else {
				g.o.SendFunc(func(o *base.Object) {
					defer atomic.AddInt64(&num, -1)
					f()
				})
			}
		}()
		defer tools.RecoverPanicFunc("goroutines error")
		if callFunc != nil {
			callFunc()
		}
	}()
}

// Go 启动一个协程，在默认节点上执行回调方法
// callFunc 在协程中执行的方法
// callbackFunc 回调方法
func Go(callFunc func(), callbackFunc ...func()) {
	New().Go(callFunc, callbackFunc...)
}

type _go struct {
	callFunc     func() // 执行方法
	callbackFunc func() // 回调方法
}

// q 协程队列，同一个队列中的协程串行执行
type q struct {
	o  *base.Object
	l  *list.List
	lm sync.Mutex
	gm sync.Mutex
}

// NewQ 创建协程队列
// o 回调方法执行节点
func NewQ(o ...*base.Object) *q {
	var obj *base.Object
	if len(o) > 0 {
		obj = o[0]
	}
	if obj == nil {
		obj = gObject
	}
	return &q{
		o: obj,
		l: list.New(),
	}
}

// Go 启动一个协程
// callFunc 在协程中执行的方法
// callbackFunc 回调方法
func (q *q) Go(callFunc func(), callbackFunc ...func()) {
	var f func()
	if len(callbackFunc) > 0 {
		f = callbackFunc[0]
	}

	atomic.AddInt64(&num, 1)

	q.lm.Lock()
	q.l.PushBack(&_go{callFunc: callFunc, callbackFunc: f})
	q.lm.Unlock()

	go func() {
		q.gm.Lock()
		defer q.gm.Unlock()

		q.lm.Lock()
		g := q.l.Remove(q.l.Front()).(*_go)
		q.lm.Unlock()

		defer func() {
			if q.o == nil || g.callbackFunc == nil {
				atomic.AddInt64(&num, -1)
			} else {
				q.o.SendFunc(func(o *base.Object) {
					defer atomic.AddInt64(&num, -1)
					g.callbackFunc()
				})
			}
		}()
		defer tools.RecoverPanicFunc("goroutines error")
		if g.callFunc != nil {
			g.callFunc()
		}
	}()
}
