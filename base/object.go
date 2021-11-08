package base

import (
	"fmt"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/container/queue"
	"github.com/skeletongo/cube/tools"
)

// Object 基础节点，单线程模型
// 包含一个消息队列及定时器，在单线程中串行处理消息队列中的所有消息及定时任务
// 优先处理队列消息，队列消息处理后查看定时任务是否需要执行
// 注意此线程中执行的方法不应该是耗时操作，否则整个线程会被阻塞
type Object struct {
	// Name 节点名称
	Name string

	// Data 节点数据
	Data interface{}

	// Opt 节点配置
	Opt *Options

	// Closing 是否关闭中
	Closing chan struct{}

	// Closed 节点是否已经关闭
	Closed chan struct{}

	// doneNum 已处理消息数量
	doneNum uint64

	// sendNum 收到的消息总数
	sendNum uint64

	// q 消息队列
	q Queue

	// signal 收到新消息的信号
	// 作用：当消息队列为空时，阻塞当前节点所在的协程，当收到新消息后不再阻塞
	signal chan struct{}

	// sinker 节点生命周期
	sinker Sinker
}

// NewObject 创建节点
// id 节点ID
// name 节点名称
// opt 节点配置
// sinker 节点生命周期
func NewObject(name string, opt *Options, sinker Sinker) *Object {
	if opt == nil {
		log.Panicf("new object error: required Options, name %s", name)
		return nil
	}
	log.Tracef("new object, name %s", name)
	opt.Init()
	o := &Object{
		Name:    name,
		Opt:     opt,
		Closing: make(chan struct{}),
		Closed:  make(chan struct{}),
		q:       queue.NewSyncQueue(),
		signal:  make(chan struct{}, 1),
		sinker:  sinker,
	}
	return o
}

// State 获取节点状态
func (o *Object) State() *State {
	return &State{
		QueueLen:   uint64(o.q.Len()),
		EnqueueNum: atomic.LoadUint64(&o.sendNum),
		DoneNum:    atomic.LoadUint64(&o.doneNum),
	}
}

// Run 启动节点
// 创建一个协程来处理消息队列中的消息和定时任务
func (o *Object) Run() {
	log.Tracef("object run, name %s", o.Name)
	o.safeStart()
	if o.Opt.Interval > 0 && o.sinker != nil {
		go o.runTicker()
	} else {
		go o.run()
	}
}

func (o *Object) runTicker() {
	t := time.NewTicker(o.Opt.Interval)
	defer t.Stop()

	for !o.canStop() {
		if o.q.Len() <= 0 {
			select {
			case <-o.signal:
			case <-t.C:
				o.safeTick()
			}
			continue
		}
		o.safeDone(o.q.Dequeue().(Command))
		select {
		case <-t.C:
			o.safeTick()
		default:
		}
	}

	o.safeStop()
	log.Tracef("object close, name %s", o.Name)
	close(o.Closed)
}

func (o *Object) run() {
	for !o.canStop() {
		if o.q.Len() <= 0 {
			<-o.signal
			continue
		}
		o.safeDone(o.q.Dequeue().(Command))
	}

	o.safeStop()
	log.Tracef("object close, name %s", o.Name)
	close(o.Closed)
}

// canStop 判定节点是否可以关闭
// 关闭条件：所有收到的消息已经处理
func (o *Object) canStop() bool {
	select {
	case <-o.Closing:
		return atomic.LoadUint64(&o.sendNum) == o.doneNum
	default:
		return false
	}
}

func (o *Object) safeDone(cmd Command) {
	defer tools.RecoverPanicFunc(fmt.Sprintf("object(%s) safeDone", o.Name))

	defer func() { atomic.AddUint64(&o.doneNum, 1) }()
	cmd.Done(o)
}

func (o *Object) safeStart() {
	defer tools.RecoverPanicFunc(fmt.Sprintf("object(%s) safeStart", o.Name))

	if o.sinker != nil {
		o.sinker.OnStart()
	}
}

func (o *Object) safeTick() {
	defer tools.RecoverPanicFunc(fmt.Sprintf("object(%s) safeTick", o.Name))

	if o.sinker != nil {
		o.sinker.OnTick()
	}
}

func (o *Object) safeStop() {
	defer tools.RecoverPanicFunc(fmt.Sprintf("object(%s) safeStop", o.Name))

	if o.sinker != nil {
		o.sinker.OnStop()
	}
}

// SendCommand 给当前节点发送消息
// 此方法为非阻塞方法，消息为异步处理，消息先进入消息队列等待处理
func (o *Object) SendCommand(c Command) {
	atomic.AddUint64(&o.sendNum, 1)
	o.q.Enqueue(c)
	select {
	case o.signal <- struct{}{}:
	default:
	}
}

func (o *Object) SendFunc(f func(o *Object)) {
	o.SendCommand(CommandWrapper(f))
}

// IsClosing 是否正在关闭
//func (o *Object) IsClosing() bool {
//	select {
//	case <-o.Closed:
//		return false
//	case <-o.Closing:
//		return true
//	default:
//		return false
//	}
//}

// IsClosed 是否已经关闭
//func (o *Object) IsClosed() bool {
//	select {
//	case <-o.Closed:
//		return true
//	default:
//		return false
//	}
//}

// Close 关闭节点
func (o *Object) Close() {
	select {
	case <-o.Closing:
		return
	case <-o.Closed:
		return
	default:
		close(o.Closing)
		// 当队列为空时，发送一个空消息，使节点立刻关闭
		o.SendCommand(new(NilCommand))
	}
}
