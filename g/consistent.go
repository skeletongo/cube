package g

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"stathat.com/c/consistent"

	"github.com/skeletongo/cube/base"
)

// Consistent  基于一致性hash，根据名称哈希值分配是否在同一个协程中执行
type Consistent struct {
	*consistent.Consistent
	*Key
	name string
}

func (c *Consistent) Go(key string, callFunc func(ctx context.Context), callbackFunc ...func()) {
	name, err := c.Get(key)
	if err != nil {
		log.Errorf("consistent get key %s error %s", key, err.Error())
		return
	}
	c.Key.Go(fmt.Sprintf("%s/%s/%s", c.name, name, key), callFunc, callbackFunc...)
}

func NewConsistent(n int, name string, o ...*base.Object) *Consistent {
	var obj *base.Object
	if len(o) > 0 {
		obj = o[0]
	}
	if obj == nil {
		obj = object
	}
	c := &Consistent{
		Consistent: consistent.New(),
		Key:        NewKey("Consistent", obj),
		name:       name,
	}
	for i := 0; i < n; i++ {
		c.Add(fmt.Sprintf("node%d", i))
	}
	return c
}

var globalConsistent *Consistent

func GoConsistent(key string, callFunc func(ctx context.Context), callbackFunc ...func()) {
	if globalConsistent == nil {
		globalConsistent = NewConsistent(Config.ConsistentNum, "GlobalConsistent")
	}
	globalConsistent.Go(key, callFunc, callbackFunc...)
}
