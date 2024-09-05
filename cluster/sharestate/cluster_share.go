package sharestate

import (
	"fmt"

	"simds-standalone/cluster/base"
	"simds-standalone/cluster/lib"
	"simds-standalone/config"
)

const shareSchdulerNum = 3

func BuildShareStateCluster() base.Cluster {

	var cluster base.Cluster
	taskgen0 := lib.NewTaskGen("simds-taskgen0")
	storage := NewStateStorage("simds-storage")

	for i := 0; i < shareSchdulerNum; i++ {
		scheduler := lib.NewCenterScheduler(fmt.Sprintf("simds-scheduler%d", i))
		scheduler.Storage = storage.GetAddress()

		cluster.Join(base.NewNode(scheduler))
		taskgen0.Receivers = append(taskgen0.Receivers, scheduler.GetAddress())
		storage.Schedulers = append(storage.Schedulers, scheduler.GetAddress())
	}

	for i := 0; i < int(config.Val.NodeNum); i++ {
		workerName := fmt.Sprintf("simds-worker%d", i)
		newworker := lib.NewWorker(
			workerName,
			lib.NodeInfo{Addr: workerName, Cpu: config.Val.NodeCpu, Memory: config.Val.NodeMemory, CpuAllocted: 0, MemoryAllocted: 0},
			storage.GetAddress(),
		)
		storage.Workers[workerName] = newworker.Node.Clone()
		cluster.Join(base.NewNode(newworker))
	}

	cluster.Join(base.NewNode(storage))
	cluster.Join(base.NewNode(taskgen0))

	return cluster

}
