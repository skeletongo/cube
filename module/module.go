package module

import (
	"container/list"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/base"
	"github.com/skeletongo/cube/utils"
)

// Module 扩展模块，实现自己的功能
type Module interface {
	// Name 模块名称
	Name() string
	// Init 模块初始化方法
	Init()
	// Update 模块更新方法
	Update()
	// Close 模块关闭方法,关闭后需要调用一下 Closed(m Module) 方法确认扩展模块已经关闭
	Close()
}

type module struct {
	// 最后一次更新时间
	lastTime time.Time
	// 更新时间间隔
	interval time.Duration
	// 优先级，越小优先级越高
	priority int
	// implement Module
	mi Module
}

func (m *module) safeInit() {
	defer utils.RecoverPanicFunc(fmt.Sprintf("module(%v) safeInit", m.mi.Name()))
	m.mi.Init()
}

func (m *module) safeUpdate(t time.Time) {
	defer utils.RecoverPanicFunc(fmt.Sprintf("module(%v) safeUpdate", m.mi.Name()))
	if m.interval == 0 || t.Sub(m.lastTime) >= m.interval {
		m.lastTime = t
		m.mi.Update()
	}
}

func (m *module) safeClose() {
	defer utils.RecoverPanicFunc(fmt.Sprintf("module(%v) safeClose", m.mi.Name()))
	m.mi.Close()
}

// 模块管理器状态
const (
	StateInvalid = iota // 停止
	StateInit           // 初始化
	StateUpdate         // 运行中
	StateClose          // 开始关闭
	StateClosing        // 关闭中
	StateClosed         // 已关闭
)

// moduleMgr 模块管理器
type moduleMgr struct {
	// state 模块管理器状态
	state int
	// mods 所有模块
	mods *list.List
	// modSign 接收模块关闭信号
	modSign chan string
	// t 定时输出还有哪些模块没有关闭
	t       <-chan time.Time
	Closing bool
	Closed  chan struct{}
}

func (m *moduleMgr) onTick() {
	switch m.state {
	case StateInit:
		m.init()
	case StateUpdate:
		m.update()
	case StateClose:
		m.close()
	case StateClosing:
		m.closing()
	case StateClosed:
		m.closed()
	}
}

func (m *moduleMgr) init() {
	log.Infoln("module init...")
	for e := m.mods.Front(); e != nil; e = e.Next() {
		mod := e.Value.(*module)
		log.Infof("module [%16s] init...", mod.mi.Name())
		mod.safeInit()
		log.Infof("module [%16s] init[ok]", mod.mi.Name())
	}
	log.Infoln("module init[ok]")

	m.state = StateUpdate
}

func (m *moduleMgr) update() {
	nowTime := time.Now()
	for e := m.mods.Front(); e != nil; e = e.Next() {
		if mod := e.Value.(*module); mod.interval > 0 {
			mod.safeUpdate(nowTime)
		}
	}
}

func (m *moduleMgr) close() {
	m.modSign = make(chan string, m.mods.Len())

	log.Infoln("module close...")
	for e := m.mods.Back(); e != nil; e = e.Prev() {
		mod := e.Value.(*module)
		log.Infof("module [%16s] close...", mod.mi.Name())
		mod.safeClose()
		log.Infof("module [%16s] close[ok]", mod.mi.Name())
	}
	log.Infoln("module close[ok]")

	m.state = StateClosing

	m.t = time.Tick(time.Second)
}

func (m *moduleMgr) closing() {
	for {
		select {
		case name := <-m.modSign:
			for e := m.mods.Front(); e != nil; e = e.Next() {
				if e.Value.(*module).mi.Name() == name {
					m.mods.Remove(e)
					break
				}
			}
		case <-m.t:
			if m.mods.Len() > 0 {
				var names []string
				for e := m.mods.Front(); e != nil; e = e.Next() {
					names = append(names, e.Value.(*module).mi.Name())
				}
				log.Info("module closing ", strings.Join(names, "|"))
			}
		default:
			if m.mods.Len() == 0 {
				m.state = StateClosed
			} else {
				m.update()
			}
			return
		}
	}
}

func (m *moduleMgr) closed() {
	m.state = StateInvalid
	log.Infoln("Module closed")
	close(m.Closed)
}

func (m *moduleMgr) Start() {
	m.state = StateInit
}

func (m *moduleMgr) Close() {
	if m.Closing {
		return
	}
	m.Closing = true
	m.state = StateClose
}

func newModuleMgr() *moduleMgr {
	ret := &moduleMgr{
		state:  StateInvalid,
		mods:   list.New(),
		Closed: make(chan struct{}),
	}
	return ret
}

var gModuleMgr = newModuleMgr()

func Closed(m Module) {
	gModuleMgr.modSign <- m.Name()
}

// Register 模块注册
// interval 执行update的时间间隔
// priority 优先级；值越小越优先处理
func Register(m Module, interval time.Duration, priority int) {
	mod := &module{
		lastTime: time.Now(),
		interval: interval,
		priority: priority,
		mi:       m,
	}
	for e := gModuleMgr.mods.Front(); e != nil; e = e.Next() {
		if me, ok := e.Value.(*module); ok {
			if priority < me.priority {
				gModuleMgr.mods.InsertBefore(mod, e)
				return
			}
		}
	}
	gModuleMgr.mods.PushBack(mod)
}

// Start 启动模块
func Start() {
	Obj.Send(base.CommandWrapper(func(o *base.Object) error {
		gModuleMgr.Start()
		return nil
	}))
}

// Close 停止所有模块
func Close() {
	Obj.Send(base.CommandWrapper(func(o *base.Object) error {
		gModuleMgr.Close()
		return nil
	}))
	<-gModuleMgr.Closed
}
