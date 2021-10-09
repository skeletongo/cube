package network

import (
	"errors"
	"fmt"
)

type Opportunity int

const (
	AfterConnected Opportunity = iota
	AfterClosed
	BeforeReceived
	AfterReceived
	BeforeSend
	AfterSend
	MaxOpportunity
)

type Filter interface {
	Get(op Opportunity) func(c *Context) bool
}

type FilterChain struct {
	functions [][]func(c *Context) bool
}

func (fc *FilterChain) Fire(op Opportunity, c *Context) bool {
	for _, f := range fc.functions[op] {
		if !f(c) {
			return false
		}
	}
	return true
}

type Middle interface {
	Get(op Opportunity) func(c *Context)
}

type MiddleChain struct {
	functions [][]func(c *Context)
}

func (m *MiddleChain) Fire(op Opportunity, c *Context) {
	for _, f := range m.functions[op] {
		f(c)
	}
}

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

func (m *FilterMgr) RegisterFilter(name string, f func() Filter) {
	m.filterCreators[name] = f
}

func (m *FilterMgr) RegisterMiddle(name string, f func() Middle) {
	m.middleCreators[name] = f
}

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

func (m *FilterMgr) AddFilter(f func() Filter) {
	m.filterChain = append(m.filterChain, f())
}

func (m *FilterMgr) AddMiddle(f func() Middle) {
	m.middleChain = append(m.middleChain, f())
}

var gFilterMgr = NewFilerMgr()

func RegisterFilter(name string, f func() Filter) {
	gFilterMgr.RegisterFilter(name, f)
}

func RegisterMiddle(name string, f func() Middle) {
	gFilterMgr.RegisterMiddle(name, f)
}

// AddFilter 添加过滤器，配置文件中的FilterChain会覆盖代码中添加的过滤器
func AddFilter(f func() Filter) {
	gFilterMgr.AddFilter(f)
}

// AddMiddle 添加中间件，配置文件中的MiddleChain会覆盖代码中添加的中间件
func AddMiddle(f func() Middle) {
	gFilterMgr.AddMiddle(f)
}

type FilterFunc struct {
	AfterConnected, BeforeReceived, AfterReceived, BeforeSend, AfterSend, AfterClosed func(c *Context) bool
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
	}
	return nil
}

type MiddleFunc struct {
	AfterConnected, BeforeReceived, AfterReceived, BeforeSend, AfterSend, AfterClosed func(c *Context)
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
	}
	return nil
}
