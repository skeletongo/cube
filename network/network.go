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

// IService 网络服务
type IService interface {
	Start() error
	Update()
	Shutdown()
}

// Network 网络服务管理器
type Network struct {
	service  map[int]IService
	configCh chan *ServiceConfig
	close    bool
}

func NewNetwork() *Network {
	return &Network{
		service:  make(map[int]IService, Capacity),
		configCh: make(chan *ServiceConfig, Capacity),
	}
}

var gNetwork = NewNetwork()

func (n *Network) Name() string {
	return "network"
}

func (n *Network) newService(config *ServiceConfig) IService {
	if n.close || config == nil {
		return nil
	}

	var s IService
	if config.IsClient {
		switch config.Protocol {
		case "ws", "wss":

		case "udp":

		default:
			s = NewTCPClient(n, config)
		}
	} else {
		switch config.Protocol {
		case "ws", "wss":

		case "udp":

		default:
			s = NewTCPServer(n, config)
		}
	}

	if s == nil {
		return nil
	}

	if err := s.Start(); err != nil {
		log.WithField("service", config.ServerKey.String()).Errorf("network service start error: %v", err)
		return nil
	}
	n.service[config.GetIndex()] = s
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
		_, ok := n.service[config.GetIndex()]
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
		module.Closed(n)
		return
	}

	for _, v := range n.service {
		v.Shutdown()
	}
}

func (n *Network) ServiceClosed(config *ServiceConfig) {
	delete(n.service, config.GetIndex())
	if !n.close {
		time.AfterFunc(TimeRestart, func() {
			n.NewService(config)
		})
		return
	}
	if len(n.service) == 0 {
		module.Closed(n)
	}
}

// goroutine safe
func (n *Network) NewService(config *ServiceConfig) {
	select {
	case gNetwork.configCh <- config:
	default:
		log.Errorf("Network: service channel full, retrying in %v", TimeRestart)
		time.AfterFunc(TimeRestart, func() {
			n.NewService(config)
		})
	}
}

// goroutine safe
func NewService(config *ServiceConfig) {
	gNetwork.NewService(config)
}
