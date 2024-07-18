package cube

import (
	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/tools"
)

var Config = tools.GetViper("config")

// Package 功能模块
type Package interface {
	// Name 模块名称
	Name() string
	// Init 初始化方法
	Init() error
	// Close 关闭方法
	Close() error
}

var packages = make(map[string]Package)

// Register 注册模块
func Register(p Package) {
	packages[p.Name()] = p
}

// Load 加载功能模块
// filePath 配置文件路径
func Load() {
	var err error
	for name := range Config.AllSettings() {
		pkg, ok := packages[name]
		if !ok {
			log.Warnf("Package %v init data not exist.", name)
			continue
		}
		if err = Config.UnmarshalKey(name, pkg); err != nil {
			log.Errorf("Unmarshalling from config file error:%s", err)
			continue
		}
		if err = pkg.Init(); err != nil {
			log.Errorf("Initializing Package %s error:%s", pkg.Name(), err)
			continue
		}
		log.Infof("Package [%16s] load success", pkg.Name())
	}
}

// Close 关闭功能模块
func Close() {
	for _, v := range packages {
		if err := v.Close(); err != nil {
			log.Errorf("Closing package %s error: %s", v.Name(), err)
		} else {
			log.Infof("Package [%16s] close success", v.Name())
		}
	}
}
