package timer

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/skeletongo/cube/base"
)

// 延时函数默认执行节点
var defaultObject *base.Object

// SetObject 设置定时器延时函数默认执行节点
func SetObject(o *base.Object) {
	defaultObject = o
}

type Handle uint32

// handles 保存所有未超时的定时器
var handles = new(sync.Map)

var i uint32

// getHandle 获取定时任务的id
func getHandle() Handle {
	return Handle(atomic.AddUint32(&i, 1))
}

func newTimer(o *base.Object, h Handle, interval time.Duration, f func()) *time.Timer {
	if o == nil {
		o = defaultObject
	}
	t := time.AfterFunc(interval, func() {
		handles.Delete(h)
		SendTimer(o, f)
	})
	handles.Store(h, t)
	return t
}

// NewTimer 创建延时方法
// o 方法执行节点，为nil时在默认节点上执行
// interval 延时时长
// f 方法实例
// 返回延时方法的id,用来提前终止执行
func NewTimer(o *base.Object, interval time.Duration, f func()) Handle {
	var h = getHandle()
	newTimer(o, h, interval, f)
	return h
}

// AfterTimer 创建在默认节点上执行的延时方法
// interval 延时时长
// f 方法实例
// 返回延时方法的id,用来提前终止执行
func AfterTimer(interval time.Duration, f func()) Handle {
	return NewTimer(defaultObject, interval, f)
}

func newCron(o *base.Object, h Handle, cronExpr *CronExpr, f func()) *time.Timer {
	now := time.Now()
	nextTime := cronExpr.Next(now)
	if nextTime.IsZero() {
		return nil
	}

	// callback
	var t *time.Timer
	var _cb func()
	_cb = func() {
		defer f()

		now := time.Now()
		nextTime := cronExpr.Next(now)
		if nextTime.IsZero() {
			return
		}
		t = newTimer(o, h, nextTime.Sub(now), _cb)
	}

	t = newTimer(o, h, nextTime.Sub(now), _cb)
	return t
}

// NewCron 创建循环定时方法
// o 方法执行节点，为nil时在默认节点上执行
// expr 定时执行规则
// f 定时执行的方法
// 返回延时方法的id,用来提前终止执行,和expr配置错误
func NewCron(o *base.Object, expr string, f func()) (Handle, error) {
	s, err := NewCronExpr(expr)
	if err != nil {
		return 0, err
	}
	var h = getHandle()
	t := newCron(o, h, s, f)
	handles.Store(h, t)
	return h, nil
}

// StartCron
// expr 定时执行规则
// f 定时执行的方法
// 返回延时方法的id,用来提前终止执行,和expr配置错误
func StartCron(expr string, f func()) (Handle, error) {
	return NewCron(defaultObject, expr, f)
}

// Stop 停止延时方法执行
func Stop(h Handle) {
	v, ok := handles.Load(h)
	if !ok {
		return
	}
	handles.Delete(h)
	v.(*time.Timer).Stop()
}

// StopAll 停止所有延时方法的执行
func StopAll() {
	handles.Range(func(key, value interface{}) bool {
		value.(*time.Timer).Stop()
		return true
	})
	handles = new(sync.Map)
}
