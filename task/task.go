// 对多线程的支持
// 主要有以下几个方法
// Go：创建一个协程去执行，执行结束后协程关闭
// GoByExecutor：在预创建的协程节点中执行
// GoByFixExecutor：创建一个协程去执行，协程一旦创建就不会关闭
package task

import "github.com/skeletongo/cube/base"

// gObject 回调方法执行节点
var gObject *base.Object

func SetObject(o *base.Object) {
	gObject = o
}

// Task 任务，需要在协程中处理的方法，通常是一些耗时的需要异步执行的操作
type Task struct {
	o            *base.Object // 回调方法执行节点
	callFunc     func()       // 执行方法
	callbackFunc func()       // 回调方法
}

func (t *Task) call() {
	if t.callFunc != nil {
		t.callFunc()
	}
	if t.callbackFunc != nil {
		sendCallback(t.o, t)
	}
}

func (t *Task) callback() {
	if t.callbackFunc != nil {
		t.callbackFunc()
	}
}

// New 创建任务
// o 回调方法执行的节点
// callFunc 需要并发执行的方法
// callBackFunc 回调方法
// name 任务名称
func New(o *base.Object, callFunc, callbackFunc func()) *Task {
	task := &Task{
		o:            o,
		callFunc:     callFunc,
		callbackFunc: callbackFunc,
	}
	if o == nil {
		task.o = gObject
	}
	return task
}

// Go 创建一个协程去执行，执行结束后协程关闭
func (t *Task) Go() {
	go t.call()
}

// GoByExecutor 在预创建的协程节点中执行
// name 标识，标识相同的任务会在同一个协程中串行执行
func (t *Task) GoByExecutor(key string) {
	sendToExecutor(t, key)
}

// GoByFixExecutor 根据标识创建一个协程去执行，协程一旦创建就不会关闭
// 如果已经有标识相同的协程了(已经使用相同的name调用过此方法)，就不会再创建新协程
// name 标识，标识相同的任务会在同一个协程中串行执行
func (t *Task) GoByFixExecutor(key string) {
	sendToFixExecutor(t, key)
}

func Go(callFunc func(), callbackFunc ...func()) *Task {
	var f func()
	if len(callbackFunc) > 0 {
		f = callbackFunc[0]
	}
	t := New(nil, callFunc, f)
	t.Go()
	return t
}

func GoByExecutor(key string, callFunc func(), callbackFunc ...func()) *Task {
	var f func()
	if len(callbackFunc) > 0 {
		f = callbackFunc[0]
	}
	t := New(nil, callFunc, f)
	t.GoByExecutor(key)
	return t
}

func GoByFixExecutor(key string, callFunc func(), callbackFunc ...func()) *Task {
	var f func()
	if len(callbackFunc) > 0 {
		f = callbackFunc[0]
	}
	t := New(nil, callFunc, f)
	t.GoByFixExecutor(key)
	return t
}
