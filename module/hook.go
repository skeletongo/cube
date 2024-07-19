package module

import log "github.com/sirupsen/logrus"

type HookType int

const (
	HookBeforeModuleInit HookType = iota // 模块初始化之前
	HookAfterModuleInit                  // 模块初始化之后
	HookBeforeModuleStop                 // 模块停止之前
	HookAfterModuleStop                  // 模块停止之后
	HookMax
)

type HookFunc func() error

var hooks [HookMax][]HookFunc

func RegisterHook(hookType HookType, f HookFunc) {
	if hookType < 0 || hookType > HookMax {
		return
	}
	hooks[hookType] = append(hooks[hookType], f)
}

func ExecuteHook(hookType HookType) error {
	if hookType < 0 || hookType > HookMax {
		return nil
	}
	log.Infof("execute hook: %d", hookType)
	var err error
	for _, h := range hooks[hookType] {
		err = h()
		if err != nil {
			return err
		}
	}
	return nil
}
