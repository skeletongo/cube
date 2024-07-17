package g

import (
	"context"
	"fmt"
	"sync"

	"github.com/skeletongo/cube/base"
)

type Group struct {
	o    *base.Object
	key  *Key
	cs   *Consistent
	name string
}

func (g *Group) Go(name string, callFunc func(ctx context.Context), callbackFunc ...func()) {
	Go(fmt.Sprintf("%s/%s", g.name, name), callFunc, callbackFunc...)
}

func (g *Group) GoKey(key string, callFunc func(ctx context.Context), callbackFunc ...func()) {
	g.key.Go(key, callFunc, callbackFunc...)
}

func (g *Group) GoConsistent(key string, callFunc func(ctx context.Context), callbackFunc ...func()) {
	g.cs.Go(key, callFunc, callbackFunc...)
}

func NewGroup(name string, o ...*base.Object) *Group {
	var obj *base.Object
	if len(o) > 0 {
		obj = o[0]
	}
	if obj == nil {
		obj = object
	}
	name = fmt.Sprintf("Group/%s", name)
	return &Group{
		o:    obj,
		name: name,
		key:  NewKey(name, obj),
		cs:   NewConsistent(Config.ConsistentNum, name, obj),
	}
}

var m = new(sync.Mutex)
var globalGroups = make(map[string]*Group)

func GetGroup(groupName string) *Group {
	m.Lock()
	defer m.Unlock()
	if g, ok := globalGroups[groupName]; ok {
		return g
	}
	g := NewGroup(groupName)
	globalGroups[groupName] = g
	return g
}
