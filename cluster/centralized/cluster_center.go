package centrailzed

import (
	"fmt"
	"simds-standalone/cluster/base"
	"simds-standalone/config"
)

func BuildCenterCluster() base.Cluster {

	var nodes []base.Node
	taskgen0 := base.NewTaskGen("simds-taskgen0")
	master0 := base.NewCenterScheduler("simds-master0")
	taskgen0.Receivers = append(taskgen0.Receivers, master0.GetHostName())
	for i := 0; i < int(config.Val.NodeNum); i++ {
		workerName := fmt.Sprintf("simds-worker%d", i)
		newworker := base.NewWorker(workerName, base.NodeInfo{workerName, config.Val.NodeCpu, config.Val.NodeMemory, 0, 0}, "simds-master0")
		master0.Workers[workerName] = newworker.Node.Clone()
		nodes = append(nodes, newworker)
	}
	nodes = append(nodes, master0)
	nodes = append(nodes, taskgen0)
	return base.Cluster{Nodes: nodes}

}
