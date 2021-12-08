package main

import (
	"fmt"
	"time"

	"github.com/skeletongo/cube"
	"github.com/skeletongo/cube/example/excel/converter"
	"github.com/skeletongo/cube/example/excel/gostruct"
	"github.com/skeletongo/cube/module"
)

type myModule struct {
}

func (m *myModule) Name() string {
	return "testModule"
}

func (m *myModule) Init() {

}

func (m *myModule) Update() {
	for _, v := range gostruct.TestSingle.Array {
		fmt.Printf("%+v\n", *v)
	}
	fmt.Println()
}

func (m *myModule) Close() {
	defer module.Release(m)
}

func main() {
	module.Register(new(myModule), time.Second*5, 0)
	cube.Register(converter.Config)
	cube.Run("config.json")
}
