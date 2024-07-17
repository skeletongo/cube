package g_test

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/base"
	"github.com/skeletongo/cube/g"
)

func init() {
	log.SetLevel(log.TraceLevel)
	g.Config.ConsistentNum = 10

	o := base.NewObject("go_test", &base.Options{}, nil)
	o.Run()
	g.SetObject(o)
}

func ExampleGoKey() {
	g.GoKey("a", func(ctx context.Context) {
		log.Info("a")
	})
	g.Close()
	// output:
	//
}

func ExampleGoConsistent() {
	g.GoConsistent("a1", func(ctx context.Context) {})
	g.GoConsistent("b2", func(ctx context.Context) {})
	g.GoConsistent("c3", func(ctx context.Context) {})
	g.GoConsistent("a1", func(ctx context.Context) {})
	g.Close()
	// output:
	//
}

func ExampleGroup_Go() {
	a := g.GetGroup("a")
	b := g.GetGroup("b")
	a.Go("a", func(ctx context.Context) {})
	b.Go("b", func(ctx context.Context) {})
	g.Close()
	// output:
	//
}

func Example() {
	g.GoKey("GoKey", func(ctx context.Context) {})
	time.Sleep(time.Second)

	g.GoConsistent("GoConsistent", func(ctx context.Context) {})
	time.Sleep(time.Second)

	a := g.GetGroup("MyGroup")
	a.Go("GoGroup", func(ctx context.Context) {})
	time.Sleep(time.Second)
	a.GoKey("groupKey", func(ctx context.Context) {})
	time.Sleep(time.Second)
	a.GoConsistent("groupConsistent", func(ctx context.Context) {})

	g.Close()
	// output:
	//
}
