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

	// 连接配置
	MTU             int           // MTU
	Linger          int           // Linger
	NoDelay         bool          // NoDelay
	KeepAlive       bool          // KeepAlive
	KeepAlivePeriod time.Duration // KeepAlivePeriod
	ReadBuffer      int           // ReadBuffer
	WriteBuffer     int           // WriteBuffer
	ReadTimeout     time.Duration // ReadTimeout
	WriteTimeout    time.Duration // WriteTimeout

	seq int
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
}

func (sc *ServiceConfig) GetSeq() int {
	sc.seq++
	return sc.seq
}

func (sc *ServiceConfig) String() string {
	return fmt.Sprintf("%s IsClient:%v Protocol:%v IP:%v Port:%v",
		sc.ServerKey.String(), sc.IsClient, sc.Protocol, sc.Ip, sc.Port)
}
