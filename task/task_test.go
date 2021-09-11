package task_test

import (
	"fmt"

	"github.com/skeletongo/cube/base"
	"github.com/skeletongo/cube/task"
)

// task 初始化
func init() {
	task.Config.Options = new(base.Options)
	task.Config.Worker = &task.WorkerConfig{
		Options:   new(base.Options),
		WorkerCnt: 5,
	}
	task.Config.Init()
}

func ExampleTask_Go() {
	ch := make(chan struct{})
	n := -1
	task.New(task.Obj, func() {
		fmt.Println("1")
		n = 2
	}, func() {
		fmt.Println(n)
		fmt.Println(3)
		ch <- struct{}{}
	}).Go()
	<-ch
	// output:
	// 1
	// 2
	// 3
}

func ExampleTask_GoByExecutor() {
	ch := make(chan struct{})
	n := -1
	task.New(task.Obj, func() {
		fmt.Println("1")
		n = 2
	}, func() {
		fmt.Println(n)
		fmt.Println(3)
		ch <- struct{}{}
	}).GoByExecutor("task")
	<-ch
	// output:
	// 1
	// 2
	// 3
}

func ExampleTask_GoByFixExecutor() {
	ch := make(chan struct{})
	n := -1
	task.New(task.Obj, func() {
		fmt.Println("1")
		n = 2
	}, func() {
		fmt.Println(n)
		fmt.Println(3)
		ch <- struct{}{}
	}).GoByFixExecutor("task")
	<-ch
	// output:
	// 1
	// 2
	// 3
}
