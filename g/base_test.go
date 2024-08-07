package g_test

import (
	"context"
	"sort"
	"testing"

	"github.com/skeletongo/cube/base"
	"github.com/skeletongo/cube/g"
)

func TestGo(t *testing.T) {
	o := base.NewObject("", &base.Options{}, nil)
	o.Run()
	g.SetObject(o)

	n := 1000
	var a int
	ch := make(chan int, n)
	for i := 0; i < n; i++ {
		v := i
		g.Go("", func(ctx context.Context) {
			ch <- v
		}, func() {
			a++
		})
	}

	g.Close()
	o.Close()
	<-o.Closed

	var arr sort.IntSlice
	for i := 0; i < n; i++ {
		arr = append(arr, <-ch)
	}
	if sort.IsSorted(arr) {
		t.Fatal("sort error")
	}
	if a != n {
		t.Fatal("serial error")
	}
}

func TestNewQ(t *testing.T) {
	o := base.NewObject("", &base.Options{}, nil)
	o.Run()

	n := 1000
	var a int
	arr := sort.IntSlice{}
	q := g.NewQ("", o)
	for i := 0; i < n; i++ {
		v := i
		q.Go(func(ctx context.Context) {
			arr = append(arr, v)
		}, func() {
			a++
		})
	}

	g.Close()
	o.Close()
	<-o.Closed

	if !sort.IsSorted(arr) {
		t.Fatal("sort error")
	}
	if a != n {
		t.Fatal("serial error")
	}
}
