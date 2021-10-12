package module

import (
	"container/list"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/base"
	"github.com/skeletongo/cube/tools"
)

// Module 自定义模块，实现应用层功能
type Module interface {
	// Name 模块名称
	Name() string

	// Init 模块初始化方法
	Init()

	// Update 模块执行方法
	// 注意此方法不能有耗时操作，否则会导致程序阻塞
	Update()

	// Close 模块关闭方法
	// 注意此方法不能有耗时操作，否则会导致程序阻塞
	// 关闭后需要调用 Release(m Module) 确认扩展模块已经关闭，否则会影响程序关闭
	Close()
}

type module struct {
	// 最后一次执行 Module.Update() 的时间
	lastTime time.Time
	// 执行 Module.Update() 的时间间隔
	interval time.Duration
	// 优先级，越小优先级越高
	priority int
	// Module
	mi Module
}

func (m *module) safeInit() {
	defer tools.RecoverPanicFunc(fmt.Sprintf("module(%v) safeInit", m.mi.Name()))

	m.mi.Init()
}

func (m *module) safeUpdate(t time.Time) {
	defer tools.RecoverPanicFunc(fmt.Sprintf("module(%v) safeUpdate", m.mi.Name()))

	if m.interval == 0 || t.Sub(m.lastTime) >= m.interval {
		m.lastTime = t
		m.mi.Update()
	}
}

func (m *module) safeClose() {
	defer tools.RecoverPanicFunc(fmt.Sprintf("module(%v) safeClose", m.mi.Name()))

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

// M 模块管理器
type M struct {
	// state 模块管理器状态
	state int

	// mods 所有模块
	mods *list.List

	// modSign 接收模块关闭信号
	modSign chan string

	// t 定时输出还有哪些模块没有关闭
	t <-chan time.Time

	isClosing bool
	Closed    chan struct{}
}

func New() *M {
	ret := &M{
		state:  StateInvalid,
		mods:   list.New(),
		Closed: make(chan struct{}),
	}
	return ret
}

func (m *M) OnTick() {
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

func (m *M) init() {
	log.Info("module init...")
	for e := m.mods.Front(); e != nil; e = e.Next() {
		mod := e.Value.(*module)
		log.Infof("module [%16s] init...", mod.mi.Name())
		mod.safeInit()
		log.Infof("module [%16s] init[ok]", mod.mi.Name())
	}
	log.Info("module init[ok]")

	m.state = StateUpdate
}

func (m *M) update() {
	nowTime := time.Now()
	for e := m.mods.Front(); e != nil; e = e.Next() {
		e.Value.(*module).safeUpdate(nowTime)
	}
}

func (m *M) close() {
	m.modSign = make(chan string, m.mods.Len())

	log.Info("module close...")
	for e := m.mods.Back(); e != nil; e = e.Prev() {
		mod := e.Value.(*module)
		log.Infof("module [%16s] close...", mod.mi.Name())
		mod.safeClose()
		log.Infof("module [%16s] close[ok]", mod.mi.Name())
	}
	log.Info("module close[ok]")

	m.state = StateClosing

	m.t = time.Tick(time.Second)
}

func (m *M) closing() {
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

func (m *M) closed() {
	m.state = StateInvalid
	log.Info("module closed")
	close(m.Closed)
}

func (m *M) Start() {
	log.Trace("module start")
	m.state = StateInit
}

func (m *M) Close() {
	log.Trace("module close")
	if m.isClosing {
		return
	}
	m.isClosing = true
	m.state = StateClose
}

// Register 注册自定义模块
// mi 自定义模块
// interval 执行 Module.Update() 的时间间隔
// priority 优先级，值越小越优先处理
func (m *M) Register(mi Module, interval time.Duration, priority int) {
	log.Tracef("module register, name %s, interval:%v, priority:%d", mi.Name(), interval, priority)
	mod := &module{
		lastTime: time.Now(),
		interval: interval,
		priority: priority,
		mi:       mi,
	}
	for e := m.mods.Front(); e != nil; e = e.Next() {
		if me, ok := e.Value.(*module); ok {
			if priority < me.priority {
				m.mods.InsertBefore(mod, e)
				return
			}
		}
	}
	m.mods.PushBack(mod)
}

// Release 确认自定义模块关闭
// 自定义模块关闭后需要主动调用此方法确认已经关闭，否则会影响程序关闭
func (m *M) Release(mod Module) {
	log.Tracef("module release, name %s", mod.Name())
	m.modSign <- mod.Name()
}

var gModuleMgr = New()

// Register 注册自定义模块
// mi 自定义模块
// interval 执行 Module.Update() 的时间间隔
// priority 优先级，值越小越优先处理
func Register(m Module, interval time.Duration, priority int) {
	gModuleMgr.Register(m, interval, priority)
}

// Release 确认自定义模块关闭
// 自定义模块关闭后需要主动调用此方法确认已经关闭，否则会影响程序关闭
func Release(m Module) {
	gModuleMgr.Release(m)
}

// Start 启动模块管理器
func Start() {
	Obj.SendFunc(func(o *base.Object) {
		gModuleMgr.Start()
	})
}

// Close 停止模块管理器
func Close() {
	Obj.SendFunc(func(o *base.Object) {
		gModuleMgr.Close()
	})
	<-gModuleMgr.Closed
}
