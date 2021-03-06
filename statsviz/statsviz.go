package statsviz

import (
	"net/http"

	"github.com/arl/statsviz"
	"github.com/sirupsen/logrus"
)

var Config = new(Configuration)

type Configuration struct {
	IsOpen bool   // 是否开启
	Addr   string // http服务地址
}

func (c *Configuration) Name() string {
	return "statsviz"
}

func (c *Configuration) Init() error {
	if c.IsOpen {
		mux := http.NewServeMux()
		if err := statsviz.Register(mux); err != nil {
			return err
		}
		go func() {
			logrus.Infof("statsviz start: %s", c.Addr)
			if err := http.ListenAndServe(c.Addr, mux); err != nil {
				logrus.Errorf("statsviz: http.ListenAndServe error: %v", err)
			}
		}()
	}
	return nil
}

func (c *Configuration) Close() error {
	return nil
}
