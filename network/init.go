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
	// Endian 字节序，默认为小端序，true表示大端序
	Endian bool
	// IsJson 修改默认编码方式为json,否则是encoding/gob
	IsJson bool
	// LenMsgLen 封包时应用层数据长度所占用的字节数
	LenMsgLen uint32
	// MinMsgLen 封包时应用层数据最短字节数
	MinMsgLen uint32
	// MaxMsgLen 封包时应用层数据最大字节数
	MaxMsgLen uint32
	// Services 网络服务配置
	Services []*ServiceConfig
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

	c.LenMsgLen, c.MinMsgLen, c.MaxMsgLen = gPkgParser.SetMsgLen(c.LenMsgLen, c.MinMsgLen, c.MaxMsgLen)

	// 启动网络服务
	module.Register(gNetwork, time.Millisecond*100, math.MaxInt32)
	return nil
}

func (c *Configuration) Close() error {
	return nil
}
