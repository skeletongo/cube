package timer_test

import (
	"fmt"
	"time"

	"github.com/skeletongo/cube/base"
	"github.com/skeletongo/cube/timer"
)

var testObj = base.NewObject("test", new(base.Options), nil)

func init() {
	testObj.Run()
}

func ExampleNewTimer() {
	ch := make(chan struct{})

	t1 := time.Now()
	timer.NewTimer(testObj, time.Second, func() {
		fmt.Println(int(time.Now().Sub(t1).Seconds()))
		ch <- struct{}{}
	})

	<-ch
	// output:
	// 1
}

func ExampleAfterTimer() {
	ch := make(chan struct{})

	t1 := time.Now()
	timer.SetObject(testObj)
	timer.AfterTimer(time.Second, func() {
		fmt.Println(int(time.Now().Sub(t1).Seconds()))
		ch <- struct{}{}
	})

	<-ch
	// output:
	// 1
}

func ExampleNewCron() {
	var t time.Time
	timer.NewCron(testObj, "*/1 * * * * *", func() {
		if t.IsZero() {
			t = time.Now()
		} else {
			now := time.Now()
			fmt.Printf("%1.0f\n", now.Sub(t).Seconds())
			t = now
		}
	})
	time.Sleep(time.Second * 3)
	// output:
	// 1
	// 1
}

func ExampleStartCron() {
	var t time.Time
	timer.SetObject(testObj)
	timer.StartCron("*/1 * * * * *", func() {
		if t.IsZero() {
			t = time.Now()
		} else {
			now := time.Now()
			fmt.Printf("%1.0f\n", now.Sub(t).Seconds())
			t = now
		}
	})
	time.Sleep(time.Second * 3)
	// output:
	// 1
	// 1
}

func ExampleStop() {
	t1 := time.Now()
	h := timer.NewTimer(testObj, time.Second, func() {
		fmt.Println(int(time.Now().Sub(t1).Seconds()))
	})

	timer.Stop(h)
	time.Sleep(time.Second * 2)

	timer.NewCron(testObj, "*/1 * * * * *", func() {
		fmt.Println("test")
	})

	timer.StopAll()
	time.Sleep(time.Second * 2)
	// output:
}
