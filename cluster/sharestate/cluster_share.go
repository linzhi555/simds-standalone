package sharestate

import (
	"fmt"
	"simds-standalone/cluster/base"
	"simds-standalone/config"
)

const shareSchdulerNum = 3

func BuildShareStateCluster() base.Cluster {

	var nodes []base.Node
	taskgen0 := base.NewTaskGen("taskgen0")
	storage := NewStateStorage("storage")

	for i := 0; i < shareSchdulerNum; i++ {
		scheduler := base.NewCenterScheduler(fmt.Sprintf("scheduler%d", i))
		scheduler.Storage = storage.GetHostName()
		nodes = append(nodes, scheduler)
		taskgen0.Receivers = append(taskgen0.Receivers, scheduler.GetHostName())
		storage.Schedulers = append(storage.Schedulers, scheduler.GetHostName())
	}

	for i := 0; i < int(config.Val.NodeNum); i++ {
		workerName := fmt.Sprintf("worker%d", i)
		newworker := base.NewWorker(workerName, base.NodeInfo{workerName, config.Val.NodeCpu, config.Val.NodeMemory, 0, 0}, storage.GetHostName())
		storage.Workers[workerName] = newworker.Node.Clone()
		nodes = append(nodes, newworker)
	}

	nodes = append(nodes, storage)
	nodes = append(nodes, taskgen0)
	return base.Cluster{Nodes: nodes}

}
