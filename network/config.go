package network

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// ServerInfo 服务标识
type ServerInfo struct {
	Area uint8  // 地区
	Type uint8  // 类型
	ID   uint16 // ID
	Name string // 名称
}

func (s *ServerInfo) String() string {
	return fmt.Sprintf("Area:%v, Type:%v, ID:%v, Name:%v", s.Area, s.Type, s.ID, s.Name)
}

// Key 服务标识转换成数字形式
func (s *ServerInfo) Key() ServerKey {
	// 由低位到高位依次 ID(16位) Type(8位)  Area(8位)
	key := uint32(0)
	key |= uint32(s.ID)
	key |= uint32(s.Area) << 16
	key |= uint32(s.Type) << 24
	return ServerKey(key)
}

// ServerKey 服务标识数字形式
type ServerKey uint32

func (s ServerKey) Parse() (areaId, typeId uint8, id uint16) {
	return uint8(s >> 24), uint8(s >> 16), uint16(s)
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	ServerInfo
	CertFile   string // 证书文件地址
	KeyFile    string // 秘钥文件地址
	Path       string // websocket连接名称
	Protocol   string // 支持的协议 "tcp" "ws" "wss"
	Ip         string // 内网ip地址
	OutIp      string // 公网ip地址
	Port       int    // 端口
	MaxRecv    int    // 接收队列缓存大小
	MaxSend    int    // 发送队列缓存大小
	MaxConnNum int    // 支持的最大连接数量（IsClient为false时有效）

	IsClient          bool          // 连接发起方
	AutoReconnect     bool          // 是否自动断线重连
	ReconnectInterval time.Duration // 重试拨号时间间隔
	ClientNum         int           // 建立连接数量（IsClient为true时有效）

	MTU             int           // 网络传输最大数据包,单位字节
	Linger          int           // 控制连接断开时的行为，连接断开后是否立刻丢弃还没有发送的缓存数据，单位秒
	KeepAlive       bool          // 是否启用tcp心跳功能
	KeepAlivePeriod time.Duration // 开启心跳功能后的发送消息的时间间隔,单位秒
	ReadBufferSize  int           // 接收数据缓冲区大小,单位字节
	WriteBufferSize int           // 发送数据缓冲区大小,单位字节
	ReadTimeout     time.Duration // 读取数据超时时长,单位秒
	WriteTimeout    time.Duration // 写入数据超时时长,单位秒
	HTTPTimeout     time.Duration // websocket 建立连接的超时时间,单位秒

	FilterChain []string     // 过滤器列表，要启用的过滤器名称及调用顺序
	filterChain *FilterChain `json:"-"`
	MiddleChain []string     // 中间件列表，要启用的中间件名称及调用顺序
	middleChain *MiddleChain `json:"-"`

	seq uint32
}

func (sc *ServiceConfig) getSeq() uint32 {
	sc.seq++
	return sc.seq
}

func (sc *ServiceConfig) init() (err error) {
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
	if sc.HTTPTimeout <= 10 {
		sc.HTTPTimeout = 10 * time.Second
	} else {
		sc.HTTPTimeout *= time.Second
	}

	if sc.filterChain, err = gFilterMgr.FilterChain(sc.FilterChain...); err != nil {
		logrus.WithField("ServiceInfo", sc).Errorf(" FilterChain error: %v", err)
		return err
	}
	if sc.middleChain, err = gFilterMgr.MiddleChain(sc.MiddleChain...); err != nil {
		logrus.WithField("ServiceInfo", sc).Errorf(" MiddleChain error: %v", err)
		return err
	}

	return err
}

func (sc *ServiceConfig) String() string {
	return fmt.Sprintf("%v, IsClient:%v, Protocol:%v, IP:%v, Port:%v",
		sc.ServerInfo.String(), sc.IsClient, sc.Protocol, sc.Ip, sc.Port)
}
