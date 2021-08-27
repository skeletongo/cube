package base

import (
	"time"
)

// Options 节点配置
type Options struct {
	Interval time.Duration // 定时任务的执行时间间隔
}

func (o *Options) Init() {
	if o.Interval > 0 {
		if o.Interval < 10 {
			o.Interval = time.Millisecond * 10
		} else {
			o.Interval = time.Millisecond * o.Interval
		}
	}
}

// State 节点状态
type State struct {
	QueueLen   uint64 // 待处理消息数量
	EnqueueNum uint64 // 收到的消息总数
	DoneNum    uint64 // 已处理的消息数
}
