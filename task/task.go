// 对多线程的支持
// 主要有以下几个方法
// Start：创建一个协程去执行，执行结束后协程关闭
// StartByExecutor：在预创建的协程节点中执行
// StartByFixExecutor：创建一个协程去执行，协程一旦创建就不会关闭
package task

import "github.com/skeletongo/cube/base"

// CallFunc 执行方法
// o 此方法执行的节点
// ret 返回值会传递给回调方法 CompleteNotify.Done()
type CallFunc func(o *base.Object) (ret interface{})

// CallbackFunc 回调方法
// ret Callable.Call() 方法的返回值
type CallbackFunc func(ret interface{}, t *Task)

// gObject 回调方法执行节点
var gObject *base.Object

func SetObject(o *base.Object) {
	gObject = o
}

// Task 任务，需要在协程中处理的方法，通常是一些耗时的需要异步执行的操作
type Task struct {
	O            *base.Object // 回调方法执行节点
	Name         string       // 任务名称
	callFunc     CallFunc     // 执行方法
	callbackFunc CallbackFunc // 回调方法
	ret          interface{}  // CallFunc 方法执行返回值
}

func (t *Task) run(o *base.Object) {
	if t.callFunc == nil {
		return
	}
	t.ret = t.callFunc(o)
	if t.callbackFunc == nil {
		return
	}
	// 在回调方法执行节点执行回调方法
	sendCallback(t.O, t)
}

// New 创建任务
// o 回调方法执行的节点
// callFunc 需要并发执行的方法
// callBackFunc 回调方法
// name 任务名称
func New(o *base.Object, callFunc CallFunc, callbackFunc CallbackFunc, name ...string) *Task {
	task := &Task{
		O:            o,
		callFunc:     callFunc,
		callbackFunc: callbackFunc,
	}
	if len(name) > 0 {
		task.Name = name[0]
	}
	if o == nil {
		task.O = gObject
	}
	return task
}

// Start 创建一个协程去执行，执行结束后协程关闭
func (t *Task) Start() {
	go t.run(nil)
}

// StartByExecutor 在预创建的协程节点中执行
// name 标识，标识相同的任务会在同一个协程中串行执行
func (t *Task) StartByExecutor(key string) {
	sendToExecutor(t, key)
}

// StartByFixExecutor 根据标识创建一个协程去执行，协程一旦创建就不会关闭
// 如果已经有标识相同的协程了(已经使用相同的name调用过此方法)，就不会再创建新协程
// name 标识，标识相同的任务会在同一个协程中串行执行
func (t *Task) StartByFixExecutor(key string) {
	sendToFixExecutor(t, key)
}
