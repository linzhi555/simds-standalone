package sharestate

import (
	"fmt"

	"simds-standalone/cluster/base"
	"simds-standalone/cluster/lib"
	"simds-standalone/config"
)

var shareSchdulerNum = int(config.Val.ShareSchdulerNum)
var storageNum = int(config.Val.StorageNum)

// sharestate cluster:
// the cluster's state is stored by some StateStorage
// each StateStorage has part of ClusterState
// each schedulers have all the Cluster State(not real time)
// if scheduler want to run a task, he should send request to Storage
// Storage will check if the Worker has enough resource.
// If yes, Storage will redirect the message to worker,If no, return CommitFail Msg
// The worker wiil send self info to StateStorage
func BuildShareStateCluster() base.Cluster {
	var cluster base.Cluster

	taskgen0 := lib.NewTaskGen("simds-taskgen0")
	var storages []*StateStorage

	// we need some storage to stora the cluster state
	for i := 0; i < storageNum; i++ {
		storages = append(storages, NewStateStorage(fmt.Sprintf("simds-storage%d", i)))
	}

	// we need some scheduler to schedule the task
	// the taskgen will send tasks to these schedulers
	// the storages will send wokerinfo to these schedulers
	for i := 0; i < shareSchdulerNum; i++ {
		scheduler := lib.NewCenterScheduler(fmt.Sprintf("simds-scheduler%d", i))
		cluster.Join(base.NewNode(scheduler))

		taskgen0.Receivers = append(taskgen0.Receivers, scheduler.GetAddress())
		for _, strg := range storages {
			strg.Schedulers = append(strg.Schedulers, scheduler.GetAddress())
		}
	}

	// we need some workers to actual execute the tasks
	// these workers should send nodeinfo to their own manager(one of the storage)
	for i := 0; i < int(config.Val.NodeNum); i++ {
		workerName := fmt.Sprintf("simds-worker%d", i)
		newworker := lib.NewWorker(
			workerName,
			lib.NodeInfo{Addr: workerName, Cpu: config.Val.NodeCpu, Memory: config.Val.NodeMemory, CpuAllocted: 0, MemoryAllocted: 0},
			storages[i%storageNum].GetAddress(), // the workerN should be managed by storage(N%storageNum)
		)
		storages[i%storageNum].Workers[workerName] = newworker.Node.Clone()
		cluster.Join(base.NewNode(newworker))
	}

	// do not forget join these nodes created before
	for _, strg := range storages {
		cluster.Join(base.NewNode(strg))
	}
	cluster.Join(base.NewNode(taskgen0))

	return cluster
}
