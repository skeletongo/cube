package network

import (
	"fmt"
	"time"
)

type ServerKey struct {
	Area int    // 地区
	Type int    // 类型
	ID   int    // ID
	Name string // 名称
}

func (s *ServerKey) String() string {
	return fmt.Sprintf("Area:%v Type:%v ID:%v Name:%v", s.Area, s.Type, s.ID, s.Name)
}

const (
	AreaBits = 1<<8 - 1
	TypeBits = 1<<8 - 1
	IDBits   = 1<<16 - 1
)

func (s *ServerKey) GetIndex() int {
	// 由低位到高位依次 ID(16位) Type(8位)  Area(8位)
	return (s.Area&AreaBits)<<24 | (s.Type&TypeBits)<<16 | s.ID&IDBits
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	ServerKey
	AuthKey    string // 秘钥
	CertFile   string // 证书文件地址
	KeyFile    string //
	Path       string // ws websocket 配置
	Protocol   string // 支持的协议 "tcp" "ws" "wss"
	Ip         string // 内网ip地址
	OutIp      string // 公网ip地址
	Port       int    // 端口
	MaxRecv    int    // 接收队列缓存大小
	MaxSend    int    // 发送队列缓存大小
	MaxConnNum int    // 支持的最大连接数量（IsClient为false时有效）

	IsClient          bool          // 连接发起方
	ReconnectInterval time.Duration // 重连间隔
	ClientNum         int           // 建立连接数量（IsClient为true时有效）

	MTU             int           // 网络传输最大数据包,单位字节
	Linger          int           // 控制连接断开时的行为，连接断开后是否立刻丢弃还没有发送的缓存数据，单位秒
	KeepAlive       bool          // 是否启用心跳功能
	KeepAlivePeriod time.Duration // 开启心跳功能后的发送消息的时间间隔,单位秒
	ReadBuffer      int           // 接收数据缓冲区大小,单位字节
	WriteBuffer     int           // 发送数据缓冲区大小,单位字节
	ReadTimeout     time.Duration // 读取数据超时时长,单位秒
	WriteTimeout    time.Duration // 写入数据超时时长,单位秒

	seq int
}

func (sc *ServiceConfig) GetSeq() int {
	sc.seq++
	return sc.seq
}

func (sc *ServiceConfig) Init() {
	if sc.MaxRecv <= 0 {
		sc.MaxRecv = 1000
	}
	if sc.MaxSend <= 0 {
		sc.MaxSend = 1000
	}
	if sc.MaxConnNum <= 0 {
		sc.MaxConnNum = 5000
	}
	if sc.ReconnectInterval < 3 {
		sc.ReconnectInterval = 3 * time.Second
	} else {
		sc.ReconnectInterval *= time.Second
	}
	if sc.ClientNum < 0 {
		sc.ClientNum = 0
	}
	if sc.KeepAlivePeriod > 0 {
		sc.KeepAlivePeriod *= time.Second
	}
	if sc.ReadTimeout > 0 {
		sc.ReadTimeout *= time.Second
	}
	if sc.WriteTimeout > 0 {
		sc.WriteTimeout *= time.Second
	}
}

func (sc *ServiceConfig) String() string {
	return fmt.Sprintf("%s IsClient:%v Protocol:%v IP:%v Port:%v",
		sc.ServerKey.String(), sc.IsClient, sc.Protocol, sc.Ip, sc.Port)
}
