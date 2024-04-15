package core

// import (
// 	"fmt"
// 	"simds-standalone/config"
// )

// // BuildDCSSCluster 建立分布式调度的集群
// // 中心化集群有三类实体 user1 任务发生器, config.Val.NodeNum 个 node 实体
// // user1 有 Taskgen组件， node实体有既有 Scheduler 也有 ResourceManager 组件（既是调度器也能worker）
// func BuildDCSSCluster() Cluster {
// 	var cluster = createCluster()
// 	var nodes []Node
// 	nodes = append(nodes, Node{
// 		"user1",
// 		[]Component{
// 			NewTaskGen("user1"),
// 		},
// 	})

// 	for i := 0; i < int(config.Val.NodeNum); i++ {

// 		nodeName := fmt.Sprintf("node%d", i)
// 		nodes = append(nodes, Node{
// 			nodeName,
// 			[]Component{
// 				NewScheduler(nodeName),
// 				NewResourceManager(nodeName),
// 			},
// 		})

// 	}
// 	cluster.Nodes = nodes
// 	cluster.RegisterFunc(CTaskGen, DcssTaskgenSetup, CommonTaskgenUpdate)
// 	cluster.RegisterFunc(CScheduler, DcssSchedulerSetup, DcssSchedulerUpdate)
// 	cluster.RegisterFunc(CResouceManger, DcssResourceManagerSetup, CommonResourceManagerUpdate)

// 	return cluster
// }
