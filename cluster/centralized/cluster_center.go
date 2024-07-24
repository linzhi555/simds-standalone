package centrailzed

import (
	"fmt"
	"simds-standalone/cluster/base"
	"simds-standalone/config"
	"simds-standalone/core"
)

func BuildCenterCluster() core.Cluster {

	var nodes []core.Node
	taskgen0 := core.NewTaskGen("taskgen0")
	master0 := base.NewCenterScheduler("master0")
	taskgen0.Receivers = append(taskgen0.Receivers, master0.GetHostName())
	for i := 0; i < int(config.Val.NodeNum); i++ {
		workerName := fmt.Sprintf("worker%d", i)
		newworker := base.NewWorker(workerName, core.NodeInfo{workerName, config.Val.NodeCpu, config.Val.NodeMemory, 0, 0}, "master0")
		master0.Workers[workerName] = newworker.Node.Clone()
		nodes = append(nodes, newworker)
	}
	nodes = append(nodes, master0)
	nodes = append(nodes, taskgen0)
	return core.Cluster{Nodes: nodes}

}
