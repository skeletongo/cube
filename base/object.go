package base

import (
	"fmt"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/container/queue"
	"github.com/skeletongo/cube/tools"
)

// Object 基础节点，单线程模型，包含一个消息队列及定时器，在自己的线程中串行处理消息队列中的所有消息及定时任务
type Object struct {
	// Name 节点名称
	Name string

	// Data 节点保存的数据
	Data interface{}

	// Opt 节点配置
	Opt *Options

	// Closing 是否关闭中
	Closing chan struct{}

	// Closed 节点是否已经关闭
	Closed chan struct{}

	// doneNum 已处理消息数量
	// sendNum 收到的消息总数
	doneNum, sendNum uint64

	// q 消息队列
	q queue.Queue

	// signal 收到新消息的信号
	// 作用：当消息队列为空时，阻塞当前节点所在的协程，当收到新消息后不再阻塞
	signal chan struct{}

	// ticker 定时器，用来定时处理定时任务
	ticker *time.Ticker

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
		log.WithField("name", name).Panicln("NewObject error: required Options")
		return nil
	}
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
	if opt.Interval > 0 && sinker != nil {
		o.ticker = time.NewTicker(opt.Interval)
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
	log.WithField("name", o.Name).Traceln("Object start")
	o.safeStart()
	go o.run()
}

func (o *Object) run() {
	// 处理队列消息及定时任务
	for !o.canStop() {
		if o.q.Len() <= 0 {
			if o.ticker == nil {
				<-o.signal
				continue
			}
			select {
			case <-o.signal:
			case <-o.ticker.C:
				o.safeTick()
			}
			continue
		}
		o.safeDone(o.q.Dequeue().(Command))
		if o.ticker != nil {
			select {
			case <-o.ticker.C:
				o.safeTick()
			default:
			}
		}
	}

	o.safeStop()
	log.WithField("name", o.Name).Traceln("Object closed")
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
	defer tools.RecoverPanicFunc(fmt.Sprintf("object(%v) safeDone", o.Name))

	defer func() { atomic.AddUint64(&o.doneNum, 1) }()

	if err := cmd.Done(o); err != nil {
		panic(err)
	}
}

func (o *Object) safeStart() {
	defer tools.RecoverPanicFunc(fmt.Sprintf("object(%v) safeStart", o.Name))

	if o.sinker != nil {
		o.sinker.OnStart()
	}
}

func (o *Object) safeTick() {
	defer tools.RecoverPanicFunc(fmt.Sprintf("object(%v) safeTick", o.Name))

	if o.sinker != nil {
		o.sinker.OnTick()
	}
}

func (o *Object) safeStop() {
	defer tools.RecoverPanicFunc(fmt.Sprintf("object(%v) safeStop", o.Name))

	if o.sinker != nil {
		o.sinker.OnStop()
	}
}

// Send 给当前节点发送消息
// 此方法为非阻塞方法，消息为异步处理，消息先进入消息队列等待处理
func (o *Object) Send(c Command) {
	atomic.AddUint64(&o.sendNum, 1)
	o.q.Enqueue(c)
	select {
	case o.signal <- struct{}{}:
	default:
	}
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
		o.Send(new(NilCommand))
	}
}
