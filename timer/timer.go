package timer

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/skeletongo/cube/base"
)

type Handle uint32

type Action interface {
	OnTimer(h Handle, ud interface{})
}

type ActionWrapper func(h Handle, ud interface{})

func (w ActionWrapper) OnTimer(h Handle, ud interface{}) {
	w(h, ud)
}

// 延时函数默认执行节点
var defaultObject *base.Object

// SetObject 设置定时器延时函数默认执行节点
func SetObject(o *base.Object) {
	defaultObject = o
}

// handles 保存所有未超时的定时器
var handles = new(sync.Map)

var i uint32

// getHandle 获取定时任务的id
func getHandle() Handle {
	return Handle(atomic.AddUint32(&i, 1))
}

type Timer struct {
	a    Action
	h    Handle
	data interface{}
}

func newTimer(o *base.Object, h Handle, a Action, data interface{}, interval time.Duration) *time.Timer {
	if o == nil {
		o = defaultObject
	}
	e := &Timer{
		a:    a,
		h:    h,
		data: data,
	}
	t := time.AfterFunc(interval, func() {
		handles.Delete(e.h)
		SendTimer(o, e)
	})
	handles.Store(h, t)
	return t
}

// NewTimer 创建延时方法
// o 方法执行节点，为nil时在默认节点上执行
// a 方法实例
// data 方法执行需要的数据
// interval 延时时长
// 返回延时方法的id,用来提前终止执行
func NewTimer(o *base.Object, a Action, data interface{}, interval time.Duration) Handle {
	var h = getHandle()
	newTimer(o, h, a, data, interval)
	return h
}

// AfterTimer 创建在默认节点上执行的延时方法
// w 方法实例
// data 方法执行需要的数据
// interval 延时时长
// 返回延时方法的id,用来提前终止执行
func AfterTimer(w ActionWrapper, data interface{}, interval time.Duration) Handle {
	return NewTimer(defaultObject, w, data, interval)
}

func newCron(o *base.Object, h Handle, cronExpr *CronExpr, cb func()) *time.Timer {
	now := time.Now()
	nextTime := cronExpr.Next(now)
	if nextTime.IsZero() {
		return nil
	}

	// callback
	var t *time.Timer
	var _cb ActionWrapper
	_cb = func(h Handle, ud interface{}) {
		defer cb()

		now := time.Now()
		nextTime := cronExpr.Next(now)
		if nextTime.IsZero() {
			return
		}
		t = newTimer(o, h, _cb, nil, nextTime.Sub(now))
	}

	t = newTimer(o, h, _cb, nil, nextTime.Sub(now))
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
