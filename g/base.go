package g

import (
	"container/list"
	"context"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/base"
	"github.com/skeletongo/cube/tools"
)

// object 回调方法默认执行节点
var object *base.Object

func SetObject(o *base.Object) {
	object = o
}

// num 启动的协程数量
var num int64

// root 用来通知所有协程关闭
var root, cancel = context.WithCancel(context.Background())

// G 等同于go协程
type G struct {
	o    *base.Object
	name string
}

// Go 启动一个协程
// callFunc 在协程中执行的方法
// callbackFunc 回调方法
func (g *G) Go(callFunc func(ctx context.Context), callbackFunc ...func()) {
	var f func()
	if len(callbackFunc) > 0 {
		f = callbackFunc[0]
	}

	atomic.AddInt64(&num, 1)

	go func() {
		defer func() {
			log.Tracef("goroutine end G/%s", g.name)
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
			log.Tracef("goroutine start G/%s", g.name)
			callFunc(root)
		}
	}()
}

// Q 协程队列，同一个队列中的协程串行执行
type Q struct {
	o    *base.Object
	l    *list.List
	lm   sync.Mutex
	gm   sync.Mutex
	name string
}

// Go 启动一个协程
// callFunc 在协程中执行的方法
// callbackFunc 回调方法
func (q *Q) Go(callFunc func(ctx context.Context), callbackFunc ...func()) {
	type _go struct {
		callFunc     func(ctx context.Context) // 执行方法
		callbackFunc func()                    // 回调方法
	}

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
			log.Tracef("goroutine end Q/%s", q.name)
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
			log.Tracef("goroutine start Q/%s", q.name)
			g.callFunc(root)
		}
	}()
}

// Close 通知所有协程关闭并等待所有协程处理完成
func Close() {
	cancel()

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

// New 创建协程对象
// o 回调方法执行节点
func New(name string, o ...*base.Object) *G {
	var obj *base.Object
	if len(o) > 0 {
		obj = o[0]
	}
	if obj == nil {
		obj = object
	}
	return &G{
		o:    obj,
		name: name,
	}
}

// Go 启动一个协程，在默认节点上执行回调方法
// callFunc 在协程中执行的方法
// callbackFunc 回调方法
func Go(name string, callFunc func(ctx context.Context), callbackFunc ...func()) {
	New(name).Go(callFunc, callbackFunc...)
}

// NewQ 创建协程队列
// o 回调方法执行节点
func NewQ(name string, o ...*base.Object) *Q {
	var obj *base.Object
	if len(o) > 0 {
		obj = o[0]
	}
	if obj == nil {
		obj = object
	}
	return &Q{
		o:    obj,
		l:    list.New(),
		name: name,
	}
}
