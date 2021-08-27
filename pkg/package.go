package pkg

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"path"

	log "github.com/sirupsen/logrus"
)

// Package 功能模块
type Package interface {
	// Name 模块名称
	Name() string
	// Load 初始化方法
	Init() error
	// Closer 关闭方法
	io.Closer
}

var packages = make(map[string]Package)

// RegisterPackage 注册模块
func RegisterPackage(p Package) {
	packages[p.Name()] = p
}

var configEncrypt Encrypt

// RegisterEncrypt 注册加解密功能
func RegisterEncrypt(h Encrypt) {
	configEncrypt = h
}

// Load 加载功能模块
// filePath 配置文件路径
func Load(filePath string) {
	switch path.Ext(filePath) {
	case ".json":
		bytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Errorf("Reading config file filepath:%s error:%s", filePath, err)
			break
		}
		if configEncrypt != nil {
			if configEncrypt.IsCipherText(bytes) {
				bytes = configEncrypt.Decode(bytes)
			}
		}
		var data interface{}
		if err = json.Unmarshal(bytes, &data); err != nil {
			log.Errorf("Reading config unmarshal filepath:%s error:%s", filePath, err)
			break
		}
		configs := data.(map[string]interface{})
		for name, pkg := range packages {
			cfg, ok := configs[name]
			if !ok {
				log.Warnf("Package %v init data not exist.", pkg.Name())
				continue
			}
			bytes, err := json.Marshal(cfg)
			if err != nil {
				log.Warnf("Package %v marshal data failed.", pkg.Name())
				continue
			}
			if err = json.Unmarshal(bytes, &pkg); err != nil {
				log.Errorf("Unmarshalling JSON from config file filepath:%s error:%s", filePath, err)
				continue
			}
			if err = pkg.Init(); err != nil {
				log.Errorf("Initializing package %s error:%s", pkg.Name(), err)
				continue
			}
			log.Infof("Package [%16s] load success", pkg.Name())
		}
	default:
		panic("Unsupported config file: " + filePath)
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
