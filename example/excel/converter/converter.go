package converter

import (
	"path/filepath"

	"github.com/howeyc/fsnotify"
	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/base"
	"github.com/skeletongo/cube/example/excel/gostruct"
	"github.com/skeletongo/cube/module"
)

// 如果直接修改目录中的excel文件很有可能会重新加载失败，原因是一个文件不能被两个程序同时访问，可以用覆盖的方式修改

var Config = new(Configuration)

type Configuration struct {
	Path string
}

func (c *Configuration) Name() string {
	return "converter"
}

func (c *Configuration) Init() error {
	if err := gostruct.Init(c.Path); err != nil {
		return err
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	p, err := filepath.Abs(c.Path)
	if err != nil {
		return err
	}
	if err := w.Watch(p); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case ev := <-w.Event:
				if ev.IsModify() || ev.IsCreate() || ev.IsRename() {
					log.Infof("watch event:%v", ev)
					module.Obj.SendFunc(func(o *base.Object) {
						log.Infof("--> load: %v", gostruct.Load(ev.Name))
					})
				}
			case err := <-w.Error:
				log.Errorf("watch error:%v", err)
			}
		}
	}()
	return nil
}

func (c *Configuration) Close() error {
	return nil
}
