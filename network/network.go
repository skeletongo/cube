package network

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/module"
)

const (
	TimeRestart = 5 * time.Second // 网络服务延迟重启时间间隔
	Capacity    = 10
)

// Service 网络服务
type Service interface {
	Start() error
	Update()
	Shutdown()
}

// Network 网络服务管理器
type Network struct {
	service  map[ServerKey]Service
	configCh chan *ServiceConfig
	close    bool
}

func NewNetwork() *Network {
	return &Network{
		service:  make(map[ServerKey]Service, Capacity),
		configCh: make(chan *ServiceConfig, Capacity),
	}
}

var gNetwork = NewNetwork()

func (n *Network) Name() string {
	return "network"
}

func (n *Network) newService(config *ServiceConfig) Service {
	if n.close || config == nil {
		return nil
	}

	var s Service
	if config.IsClient {
		switch config.Protocol {
		case "ws", "wss":
			s = NewWSClient(n, config)
		default:
			s = NewTCPClient(n, config)
		}
	} else {
		switch config.Protocol {
		case "ws", "wss":
			s = NewWSServer(n, config)
		default:
			s = NewTCPServer(n, config)
		}
	}

	if s == nil {
		log.WithField("ServiceInfo", config).Errorf("not implemented Protocol %s", config.Protocol)
		return nil
	}

	if err := s.Start(); err != nil {
		log.WithField("ServiceInfo", config).Errorf("network service start error: %v", err)
		return nil
	}
	n.service[config.Key()] = s
	return s
}

func (n *Network) Init() {
	for i := 0; i < len(Config.Services); i++ {
		n.newService(Config.Services[i])
	}
}

func (n *Network) Update() {
	select {
	case config := <-n.configCh:
		_, ok := n.service[config.Key()]
		if !n.close && !ok {
			n.newService(config)
		}

	default:
		for _, v := range n.service {
			v.Update()
		}
	}
}

func (n *Network) Close() {
	if n.close {
		return
	}
	n.close = true

	if len(n.service) == 0 {
		module.Release(n)
		return
	}

	for _, v := range n.service {
		v.Shutdown()
	}
}

// Release 确认网络服务已经关闭
// 网络服务关闭后需要主动调用此方法确认已经关闭，否则会影响程序关闭
// config 服务配置
func (n *Network) Release(config *ServiceConfig) {
	delete(n.service, config.Key())
	if !n.close {
		time.AfterFunc(TimeRestart, func() {
			n.NewService(config)
		})
		return
	}
	if len(n.service) == 0 {
		module.Release(n)
	}
}

// NewService 新增网络服务
// config 服务配置
func (n *Network) NewService(config *ServiceConfig) {
	select {
	case gNetwork.configCh <- config:
	default:
		log.Warningf("Network: service channel full, retrying in %v", TimeRestart)
		time.AfterFunc(TimeRestart, func() {
			n.NewService(config)
		})
	}
}

// NewService 新增网络服务
// config 服务配置
func NewService(config *ServiceConfig) {
	gNetwork.NewService(config)
}
