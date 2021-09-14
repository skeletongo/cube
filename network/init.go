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
	// IsJson 修改默认编码方式为json,否则是encoding/gob
	IsJson bool
}

func (c *Configuration) Name() string {
	return "network"
}

func (c *Configuration) Init() error {
	for i := 0; i < len(c.Services); i++ {
		c.Services[i].Init()
	}

	if c.Endian {
		encoding.SetEndian(binary.BigEndian)
		gMsgParser.SetByteOrder(binary.BigEndian)
		gPkgParser.SetByteOrder(binary.BigEndian)
	}

	if c.IsJson {
		encoding.SetDefaultEncodeType(encoding.TypeJson)
	}

	// 启动网络服务
	module.Register(gNetwork, time.Millisecond*100, math.MaxInt32)
	return nil
}

func (c *Configuration) Close() error {
	return nil
}
