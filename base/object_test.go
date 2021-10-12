package base_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/skeletongo/cube/base"
)

func ExampleObject_Send() {
	var n []int
	obj := base.NewObject("test", &base.Options{Interval: 0}, nil)
	obj.Run()
	obj.SendFunc(func(o *base.Object) {
		n = append(n, 1)
	})
	obj.SendFunc(func(o *base.Object) {
		fmt.Println(n)
	})
	obj.SendFunc(func(o *base.Object) {
		n = append(n, 2)
	})
	obj.SendFunc(func(o *base.Object) {
		n = append(n, 3)
	})
	obj.SendFunc(func(o *base.Object) {
		fmt.Println(n)
	})
	obj.Close()
	<-obj.Closed
	// Output:
	// [1]
	// [1 2 3]
}

var RunCh = make(chan int, 11)

type runSinker struct {
}

func (r *runSinker) OnStart() {
	RunCh <- 1
}

func (r *runSinker) OnTick() {
	RunCh <- 6
}

func (r *runSinker) OnStop() {
	RunCh <- -1
}

func TestObject_Run(t *testing.T) {
	obj := base.NewObject("test", &base.Options{Interval: 50}, new(runSinker))
	obj.Run()

	for i := 2; i < 6; i++ {
		go func(n int) {
			obj.SendFunc(func(o *base.Object) {
				RunCh <- n
			})
		}(i)
	}
	time.Sleep(time.Millisecond * 10)
	// 1 [2 3 4 5]

	obj.SendFunc(func(o *base.Object) {
		time.Sleep(time.Millisecond * 50)
	})
	// 1 [2 3 4 5] 6

	for i := 7; i < 9; i++ {
		go func(n int) {
			obj.SendFunc(func(o *base.Object) {
				RunCh <- n
				time.Sleep(time.Millisecond * 10)
			})
		}(i)
	}
	time.Sleep(time.Millisecond * 100)
	// 1 [2 3 4 5] 6 [7 8] 6

	obj.SendFunc(func(o *base.Object) {
		RunCh <- 9
	})
	// 1 [2 3 4 5] 6 [7 8] 6 9

	obj.Close()
	<-obj.Closed
	// 1 [2 3 4 5] 6 [7 8] 6 9 -1

	var res []int
	for {
		select {
		case d := <-RunCh:
			res = append(res, d)
		default:
			switch {
			case res[0] != 1 || res[5] != 6 || res[8] != 6 || res[9] != 9 || res[10] != -1:
				t.Error(res)
			case res[1]+res[2]+res[3]+res[4] != 2+3+4+5:
				t.Error(res)
			case res[6]+res[7] != 7+8:
				t.Error(res)
			}
			return
		}
	}
}
