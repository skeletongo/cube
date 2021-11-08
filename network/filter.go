package network

import (
	"errors"
	"fmt"
)

// Opportunity 过滤器及中间件的执行时机
type Opportunity int

const (
	AfterConnected Opportunity = iota // 建立连接之后
	AfterClosed                       // 连接关闭之后
	BeforeReceived                    // 消息处理之前
	AfterReceived                     // 消息处理之后
	BeforeSend                        // 发送消息之前
	AfterSend                         // 发送消息之后
	ErrorMsgID                        // 收到未注册的消息时
	MaxOpportunity
)

// Filter 过滤器
type Filter interface {
	// Get 获取特定时机的过滤方法
	Get(op Opportunity) func(c *Context) bool
}

// FilterChain 过滤器调用链
type FilterChain struct {
	functions [][]func(c *Context) bool
}

// Fire 在特定时机调用过滤器方法链
func (fc *FilterChain) Fire(op Opportunity, c *Context) bool {
	for _, f := range fc.functions[op] {
		if !f(c) {
			return false
		}
	}
	return true
}

// Middle 中间件
type Middle interface {
	// Get 获取特定时机的中间件方法
	Get(op Opportunity) func(c *Context)
}

// MiddleChain 中间件调用链
type MiddleChain struct {
	functions [][]func(c *Context)
}

// Fire 在特定时机调用中间件方法链
func (m *MiddleChain) Fire(op Opportunity, c *Context) {
	for _, f := range m.functions[op] {
		f(c)
	}
}

// FilterMgr 过滤器及中间件管理器
type FilterMgr struct {
	filterChain    []Filter
	middleChain    []Middle
	filterCreators map[string]func() Filter
	middleCreators map[string]func() Middle
}

func NewFilerMgr() *FilterMgr {
	return &FilterMgr{
		filterCreators: make(map[string]func() Filter),
		middleCreators: make(map[string]func() Middle),
	}
}

// RegisterFilter 注册过滤器
// name 名称
// f 过滤器创建方法
func (m *FilterMgr) RegisterFilter(name string, f func() Filter) {
	m.filterCreators[name] = f
}

// RegisterMiddle 注册中间件
// name 名称
// f 中间件创建方法
func (m *FilterMgr) RegisterMiddle(name string, f func() Middle) {
	m.middleCreators[name] = f
}

// FilterChain 根据名称获取过滤器调用链
// name 过滤器名称，也是多个过滤器的调用顺序
func (m *FilterMgr) FilterChain(name ...string) (chain *FilterChain, err error) {
	chain = &FilterChain{
		functions: make([][]func(c *Context) bool, MaxOpportunity),
	}
	if len(name) > 0 {
		m.filterChain = m.filterChain[:0]
		for _, v := range name {
			f, ok := m.filterCreators[v]
			if !ok {
				return nil, errors.New(fmt.Sprintf("filter not found: %s", v))
			}
			m.filterChain = append(m.filterChain, f())
		}
	}
	for _, v := range m.filterChain {
		if v == nil {
			continue
		}
		for i := 0; i < int(MaxOpportunity); i++ {
			f := v.Get(Opportunity(i))
			if f != nil {
				chain.functions[i] = append(chain.functions[i], f)
			}
		}
	}
	return
}

// MiddleChain 根据名称获取中间件调用链
// name 中间件名称，也是多个中间件的调用顺序
func (m *FilterMgr) MiddleChain(name ...string) (chain *MiddleChain, err error) {
	chain = &MiddleChain{
		functions: make([][]func(c *Context), MaxOpportunity),
	}
	if len(name) > 0 {
		m.middleChain = m.middleChain[:0]
		for _, v := range name {
			f, ok := m.middleCreators[v]
			if !ok {
				return nil, errors.New(fmt.Sprintf("middle not found: %s", v))
			}
			m.middleChain = append(m.middleChain, f())
		}
	}
	for _, v := range m.middleChain {
		if v == nil {
			continue
		}
		for i := 0; i < int(MaxOpportunity); i++ {
			f := v.Get(Opportunity(i))
			if f != nil {
				chain.functions[i] = append(chain.functions[i], f)
			}
		}
	}
	return
}

// AddFilter 追加过滤器
func (m *FilterMgr) AddFilter(f func() Filter) {
	m.filterChain = append(m.filterChain, f())
}

// AddMiddle 追加中间件
func (m *FilterMgr) AddMiddle(f func() Middle) {
	m.middleChain = append(m.middleChain, f())
}

type FilterFunc struct {
	AfterConnected, BeforeReceived, AfterReceived, BeforeSend, AfterSend, AfterClosed, ErrorMsgID func(c *Context) bool
}

func (m *FilterFunc) Get(op Opportunity) func(c *Context) bool {
	switch op {
	case AfterConnected:
		return m.AfterConnected
	case BeforeReceived:
		return m.BeforeReceived
	case AfterReceived:
		return m.AfterReceived
	case BeforeSend:
		return m.BeforeSend
	case AfterSend:
		return m.AfterSend
	case AfterClosed:
		return m.AfterClosed
	case ErrorMsgID:
		return m.ErrorMsgID
	}
	return nil
}

type MiddleFunc struct {
	AfterConnected, BeforeReceived, AfterReceived, BeforeSend, AfterSend, AfterClosed, ErrorMsgID func(c *Context)
}

func (m *MiddleFunc) Get(op Opportunity) func(c *Context) {
	switch op {
	case AfterConnected:
		return m.AfterConnected
	case BeforeReceived:
		return m.BeforeReceived
	case AfterReceived:
		return m.AfterReceived
	case BeforeSend:
		return m.BeforeSend
	case AfterSend:
		return m.AfterSend
	case AfterClosed:
		return m.AfterClosed
	case ErrorMsgID:
		return m.ErrorMsgID
	}
	return nil
}

var gFilterMgr = NewFilerMgr()

// RegisterFilter 注册过滤器
// name 名称
// f 过滤器创建方法
func RegisterFilter(name string, f func() Filter) {
	gFilterMgr.RegisterFilter(name, f)
}

// RegisterMiddle 注册中间件
// name 名称
// f 中间件创建方法
func RegisterMiddle(name string, f func() Middle) {
	gFilterMgr.RegisterMiddle(name, f)
}

// AddFilter 追加过滤器，配置文件中的FilterChain会覆盖代码中添加的过滤器
func AddFilter(f func() Filter) {
	gFilterMgr.AddFilter(f)
}

// AddMiddle 追加中间件，配置文件中的MiddleChain会覆盖代码中添加的中间件
func AddMiddle(f func() Middle) {
	gFilterMgr.AddMiddle(f)
}
