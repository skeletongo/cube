package module

import (
	"github.com/skeletongo/cube/base"
)

var Obj *base.Object

type sink struct {
}

func (s *sink) OnStart() {
}

func (s *sink) OnTick() {
	gModuleMgr.OnTick()
}

func (s *sink) OnStop() {
}

var Config = new(Configuration)

type Configuration struct {
	Options *base.Options
}

func (c *Configuration) Name() string {
	return "module"
}

func (c *Configuration) Init() error {
	Obj = base.NewObject("module", c.Options, new(sink))
	Obj.Run()
	return nil
}

func (c *Configuration) Close() error {
	return nil
}
