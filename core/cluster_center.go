package core

import (
	"fmt"
	"simds-standalone/config"
)

func BuildCenterCluster() Cluster {

	var nodes []Node
	taskgen0 := NewTaskGen("taskgen0")
	master0 := NewCenterScheduler("master0")
	taskgen0.Receivers = append(taskgen0.Receivers, master0.GetHostName())
	for i := 0; i < int(config.Val.NodeNum); i++ {
		workerName := fmt.Sprintf("worker%d", i)
		newworker := NewWorker(workerName, NodeInfo{workerName, config.Val.NodeCpu, config.Val.NodeMemory, 0, 0}, "master0")
		master0.Workers[workerName] = newworker.Node.Clone()
		nodes = append(nodes, newworker)
	}
	nodes = append(nodes, master0)
	nodes = append(nodes, taskgen0)
	return Cluster{nodes}

}
