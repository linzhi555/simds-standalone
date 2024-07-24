package sharestate

import (
	"fmt"
	"simds-standalone/cluster/base"
	"simds-standalone/config"
	"simds-standalone/core"
)

const shareSchdulerNum = 3

func BuildShareStateCluster() core.Cluster {

	var nodes []core.Node
	taskgen0 := core.NewTaskGen("taskgen0")
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
		newworker := base.NewWorker(workerName, core.NodeInfo{workerName, config.Val.NodeCpu, config.Val.NodeMemory, 0, 0}, storage.GetHostName())
		storage.Workers[workerName] = newworker.Node.Clone()
		nodes = append(nodes, newworker)
	}

	nodes = append(nodes, storage)
	nodes = append(nodes, taskgen0)
	return core.Cluster{Nodes: nodes}

}
