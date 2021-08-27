package task_test

import (
	"fmt"
	"testing"
	"time"

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

func ExampleTask_Start() {
	ch := make(chan struct{})
	task.New(task.Obj, func(o *base.Object) interface{} {
		fmt.Println("1")
		return "2"
	}, func(ret interface{}, t *task.Task) {
		fmt.Println(ret)
		fmt.Println(3)
		ch <- struct{}{}
	}).Start()
	<-ch
	// output:
	// 1
	// 2
	// 3
}

func ExampleTask_StartByExecutor() {
	ch := make(chan struct{})
	task.New(task.Obj, func(o *base.Object) interface{} {
		fmt.Println("1")
		return "2"
	}, func(ret interface{}, t *task.Task) {
		fmt.Println(ret)
		fmt.Println(3)
		ch <- struct{}{}
	}).StartByExecutor("task")
	<-ch
	// output:
	// 1
	// 2
	// 3
}

func ExampleTask_StartByFixExecutor() {
	ch := make(chan struct{})
	task.New(task.Obj, func(o *base.Object) interface{} {
		fmt.Println("1")
		return "2"
	}, func(ret interface{}, t *task.Task) {
		fmt.Println(ret)
		fmt.Println(3)
		ch <- struct{}{}
	}).StartByFixExecutor("task")
	<-ch
	// output:
	// 1
	// 2
	// 3
}

func TestTask_StartByExecutor(t *testing.T) {
	ch := make(chan string, 3)
	A, B := "a", "b"

	task.New(task.Obj, func(o *base.Object) interface{} {
		t.Logf("task name: %s, object name: %s\n", A, o.Name)
		return o.Name
	}, func(ret interface{}, t *task.Task) {
		ch <- fmt.Sprint(ret, A)
	}).StartByExecutor(A)

	task.New(task.Obj, func(o *base.Object) interface{} {
		t.Logf("task name: %s, object name: %s\n", A, o.Name)
		return o.Name
	}, func(ret interface{}, t *task.Task) {
		ch <- fmt.Sprint(ret, A)
	}).StartByExecutor(A)

	task.New(task.Obj, func(o *base.Object) interface{} {
		t.Logf("task name: %s, object name: %s\n", B, o.Name)
		return o.Name
	}, func(ret interface{}, t *task.Task) {
		ch <- fmt.Sprint(ret, B)
	}).StartByExecutor(B)

	var names []string
	for i := 0; i < 3; i++ {
		select {
		case v := <-ch:
			names = append(names, v)
		case <-time.Tick(time.Second):
			t.Error("1")
			return
		}
	}
	t.Log(names)

	if len(names) != 3 {
		t.Error("2")
		return
	}

	if names[0] == names[1] && names[0] != names[2] {
		return
	}
	if names[1] == names[2] && names[0] != names[1] {
		return
	}
	if names[0] == names[2] && names[0] != names[1] {
		return
	}
	t.Error("3")
}

func TestTask_StartByFixExecutor(t *testing.T) {
	ch := make(chan string, 3)
	A, B := "a", "b"

	task.New(task.Obj, func(o *base.Object) interface{} {
		t.Logf("task name: %s, object name: %s\n", A, o.Name)
		return o.Name
	}, func(ret interface{}, t *task.Task) {
		ch <- fmt.Sprint(ret, A)
	}).StartByFixExecutor(A)

	task.New(task.Obj, func(o *base.Object) interface{} {
		t.Logf("task name: %s, object name: %s\n", A, o.Name)
		return o.Name
	}, func(ret interface{}, t *task.Task) {
		ch <- fmt.Sprint(ret, A)
	}).StartByFixExecutor(A)

	task.New(task.Obj, func(o *base.Object) interface{} {
		t.Logf("task name: %s, object name: %s\n", B, o.Name)
		return o.Name
	}, func(ret interface{}, t *task.Task) {
		ch <- fmt.Sprint(ret, B)
	}).StartByFixExecutor(B)

	var names []string
	for i := 0; i < 3; i++ {
		select {
		case v := <-ch:
			names = append(names, v)
		case <-time.Tick(time.Second):
			t.Error("1")
			return
		}
	}
	t.Log(names)

	if len(names) != 3 {
		t.Error("2")
		return
	}

	if names[0] == names[1] && names[0] != names[2] {
		return
	}
	if names[1] == names[2] && names[0] != names[1] {
		return
	}
	if names[0] == names[2] && names[0] != names[1] {
		return
	}
	t.Error("3")
}
