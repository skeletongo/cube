package g

import (
	"context"
	"fmt"
	"sync"

	"github.com/skeletongo/cube/base"
)

// Key 相同名称的任务在同一个协程中执行
type Key struct {
	sync.Mutex
	o    *base.Object
	keys map[string]*Q
	name string
}

func (g *Key) Go(key string, callFunc func(ctx context.Context), callbackFunc ...func()) {
	g.Lock()
	defer g.Unlock()
	q, ok := g.keys[key]
	if ok {
		q.Go(callFunc, callbackFunc...)
		return
	}
	q = NewQ(fmt.Sprintf("%s/%s", g.name, key), g.o)
	g.keys[key] = q
	q.Go(callFunc, callbackFunc...)
}

// NewKey 创建协程对象
// o 回调方法执行节点
func NewKey(name string, o ...*base.Object) *Key {
	var obj *base.Object
	if len(o) > 0 {
		obj = o[0]
	}
	if obj == nil {
		obj = object
	}
	return &Key{
		o:    obj,
		keys: map[string]*Q{},
		name: name,
	}
}

var globalKey = NewKey("GlobalQueue")

func GoKey(key string, callFunc func(ctx context.Context), callbackFunc ...func()) {
	globalKey.Go(key, callFunc, callbackFunc...)
}
