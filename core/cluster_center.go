package core

// import (
// 	"fmt"
// 	"math/rand"
// 	"simds-standalone/common"
// 	"simds-standalone/config"
// )

// // BuildCenterCluster 建立中心化调度的集群
// // 中心化集群有三类实体 user1 任务发生器, master1 调度器 , config.Val.NodeNum 个worker
// // 其三类实体分别有 Taskgen Schdueler ResourceManager 组件
// // 最后注册这些组件的初始化和运行更新行为
// func BuildCenterCluster() Cluster {
// 	var cluster = createCluster()
// 	var nodes []Node
// 	nodes = append(nodes, Node{
// 		"user1",
// 		[]Component{
// 			NewTaskGen("user1"),
// 		},
// 	})

// 	nodes = append(nodes, Node{
// 		"master1",
// 		[]Component{
// 			NewScheduler("master1"),
// 		},
// 	})
// 	for i := 0; i < int(config.Val.NodeNum); i++ {

// 		workerName := fmt.Sprintf("worker%d", i)
// 		nodes = append(nodes, Node{
// 			workerName,
// 			[]Component{
// 				NewResourceManager(workerName),
// 			},
// 		})

// 	}
// 	cluster.Nodes = nodes
// 	cluster.RegisterFunc(CTaskGen, CenterTaskgenSetup, CommonTaskgenUpdate)
// 	cluster.RegisterFunc(CScheduler, CenterSchedulerSetup, CenterSchedulerUpdate)
// 	cluster.RegisterFunc(CResouceManger, CenterResourceManagerSetup, CommonResourceManagerUpdate)

// 	return cluster
// }

// // CenterTaskgenSetup 对中心化集群的任务发生器进行初始化
// // 发送对象只有一个，发送时间未测试开始时
// func CenterTaskgenSetup(c Component) {
// 	taskgen := c.(*TaskGen)
// 	taskgen.StartTime = taskgen.Os.GetTime()
// 	taskgen.Receivers = append(taskgen.Receivers, "master1"+":"+string(CScheduler))
// }

// // CenterSchedulerSetup 模拟开始时对中心化集群的调度器组件进行初始化
// // 调度器注册所有的worker信息，包括任务发送地址和节点资源信息
// func CenterSchedulerSetup(comp Component) {
// 	scheduler := comp.(*Scheduler)
// 	for i := 0; i < int(config.Val.NodeNum); i++ {

// 		nodeAddr := "worker" + fmt.Sprint(i) + ":" + string(CResouceManger)
// 		nodeinfo := &NodeInfo{nodeAddr, config.Val.NodeCpu, config.Val.NodeMemory, 0, 0}
// 		scheduler.Workers["worker"+fmt.Sprint(i)+":"+string(CResouceManger)] = nodeinfo.Clone()
// 	}

// }
