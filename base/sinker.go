package base

// Sinker 节点生命周期
type Sinker interface {
	OnStart() // 节点创建时执行
	OnTick()  // 节点协程中执行
	OnStop()  // 节点协程中执行
}
