package network

import (
	"encoding/binary"
	"math"
	"time"

	"github.com/skeletongo/cube/module"
	"github.com/skeletongo/cube/network/encoding"
)

var Config = new(Configuration)

type Configuration struct {
	// Services 网络服务配置
	Services []*ServiceConfig
	// Endian 字节序，默认为小端序，true表示大端序
	Endian bool
}

func (c *Configuration) Name() string {
	return "network"
}

func (c *Configuration) Init() error {
	if c.Endian {
		encoding.SetEndian(binary.BigEndian)
		gMsgParser.SetByteOrder(binary.BigEndian)
		gPkgParser.SetByteOrder(binary.BigEndian)
	}

	// 网络服务配置初始化
	for i := 0; i < len(c.Services); i++ {
		c.Services[i].Init()
	}

	// 启动网络服务
	module.Register(gNetwork, time.Millisecond*100, math.MaxInt32)
	return nil
}

func (c *Configuration) Close() error {
	return nil
}
