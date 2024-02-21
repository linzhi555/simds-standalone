package core

// Node 有可以有一个或多个NodeComponents
// 可以理解为一台机器上可有一个多个运行的进程
type Node struct {
	Name       string
	Components []NodeComponent
}

// Cluster 由一系列的Nodes构成
// Nodes 存储了集群的一些列组件的状态信息
// Setups 定义了集群初始化如何初始化集群的状态信息，根据组件进行初始化
// Update  定义了如何在运行时如何更新集群所有组件类型的状态性息
type Cluster struct {
	Nodes   []Node
	Setups  map[ComponentName]func(interface{})
	Updates map[ComponentName]func(interface{})
}

func createCluster() Cluster {
	return Cluster{
		Nodes:   make([]Node, 0),
		Setups:  make(map[ComponentName]func(interface{})),
		Updates: make(map[ComponentName]func(interface{})),
	}

}

// RegisterFunc 注册某个组件类型的初始化和更新方法
func (cluster *Cluster) RegisterFunc(c ComponentName, setup, update func(interface{})) {

	cluster.Setups[c] = setup
	cluster.Updates[c] = update
}
