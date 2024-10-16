package centrailzed

import (
	"fmt"
	"simds-standalone/cluster/base"
	"simds-standalone/cluster/lib"
	"simds-standalone/config"
)

func BuildCenterCluster() base.Cluster {

	var cluster base.Cluster
	taskgen0 := lib.NewTaskGen("simds-taskgen0")
	master0 := lib.NewCenterScheduler("simds-master0")
	taskgen0.Receivers = append(taskgen0.Receivers, master0.GetAddress())

	for i := 0; i < int(config.Val.NodeNum); i++ {
		workerName := fmt.Sprintf("simds-worker%d", i)
		newworker := lib.NewWorker(workerName, lib.NodeInfo{Addr: workerName, Cpu: config.Val.NodeCpu, Memory: config.Val.NodeMemory, CpuAllocted: 0, MemoryAllocted: 0}, "simds-master0")
		master0.Workers[workerName] = newworker.Node.Clone()
		cluster.Join(base.NewNode(newworker))
	}

	cluster.Join(base.NewNode(taskgen0))
	cluster.Join(base.NewNode(master0))

	return cluster

}
