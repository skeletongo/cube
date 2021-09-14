package main

import (
	"time"

	"github.com/skeletongo/cube"
	"github.com/skeletongo/cube/module"
)

type myModule struct {
}

func (m *myModule) Name() string {
	return "testModule"
}

func (m *myModule) Init() {
	panic("my module panic")
}

func (m *myModule) Update() {
	panic("my module panic")
}

func (m *myModule) Close() {
	defer module.Closed(m)
	panic("my module panic")
}

func main() {
	module.Register(new(myModule), time.Second*5, 0)
	cube.Run("config.json")
}
