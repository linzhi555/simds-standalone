package core

import (
	"fmt"
	"simds-standalone/config"
)

const shareSchdulerNum = 3

func BuildShareStateCluster() Cluster {

	var nodes []Node
	taskgen0 := NewTaskGen("taskgen0")
	storage := NewStateStorage("storage")

	for i := 0; i < shareSchdulerNum; i++ {
		scheduler := NewCenterScheduler(fmt.Sprintf("scheduler%d", i))
		scheduler.storage = storage.GetHostName()
		nodes = append(nodes, scheduler)
		taskgen0.Receivers = append(taskgen0.Receivers, scheduler.GetHostName())
		storage.Schedulers = append(storage.Schedulers, scheduler.GetHostName())
	}

	for i := 0; i < int(config.Val.NodeNum); i++ {
		workerName := fmt.Sprintf("worker%d", i)
		newworker := NewWorker(workerName, NodeInfo{workerName, config.Val.NodeCpu, config.Val.NodeMemory, 0, 0}, storage.GetHostName())
		storage.Workers[workerName] = newworker.Node.Clone()
		nodes = append(nodes, newworker)
	}

	nodes = append(nodes, storage)
	nodes = append(nodes, taskgen0)
	return Cluster{nodes}

}
