package timer_test

import (
	"fmt"
	"github.com/skeletongo/cube/base"
	"time"

	"github.com/skeletongo/cube/timer"
)

var testObj = base.NewObject("test", new(base.Options), nil)

func init() {
	testObj.Run()
}

func ExampleNewTimer() {
	ch := make(chan struct{})

	timer.NewTimer(testObj, timer.ActionWrapper(func(h timer.Handle, ud interface{}) {
		fmt.Println(int(time.Now().Sub(ud.(time.Time)).Seconds()))
		ch <- struct{}{}
	}), time.Now(), time.Second)

	<-ch
	// output:
	// 1
}

func ExampleAfterTimer() {
	ch := make(chan struct{})

	timer.SetObject(testObj)
	timer.AfterTimer(func(h timer.Handle, ud interface{}) {
		fmt.Println(int(time.Now().Sub(ud.(time.Time)).Seconds()))
		ch <- struct{}{}
	}, time.Now(), time.Second)

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
	h := timer.NewTimer(testObj, timer.ActionWrapper(func(h timer.Handle, ud interface{}) {
		fmt.Println(int(time.Now().Sub(ud.(time.Time)).Seconds()))
	}), time.Now(), time.Second)

	timer.Stop(h)
	time.Sleep(time.Second * 2)

	timer.NewCron(testObj, "*/1 * * * * *", func() {
		fmt.Println("test")
	})

	timer.StopAll()
	time.Sleep(time.Second * 2)
	// output:
}
