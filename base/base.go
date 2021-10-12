package base

// Queue 消息队列
// 需要并发安全
type Queue interface {
	Len() int
	Enqueue(interface{})
	Dequeue() interface{}
}

// Command 消息
type Command interface {
	// Done 执行方法
	Done(o *Object)
}

// NilCommand 空消息
type NilCommand struct {
}

func (n *NilCommand) Done(o *Object) {
}

type CommandWrapper func(o *Object)

func (cw CommandWrapper) Done(o *Object) {
	cw(o)
}
