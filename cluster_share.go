package main

import "fmt"

const ShareSchdulerNum = 3
const stateCopyUpdateMS = 2000

func BuildShareStateCluster() Cluster {
	var cluster = createCluster()
	var nodes []Node
	nodes = append(nodes, Node{
		"user1",
		[]NodeComponent{
			NewTaskGen("user1"),
		},
	})

	nodes = append(nodes, Node{
		"globalStateStorage",
		[]NodeComponent{
			NewStateStorage("globalStateStorage"),
		},
	})

	for i := 0; i < ShareSchdulerNum; i++ {
		masterName := fmt.Sprintf("master%d", i)
		nodes = append(nodes, Node{
			masterName,
			[]NodeComponent{
				NewScheduler(masterName),
			},
		})
	}

	for i := 0; i < int(Config.NodeNum); i++ {
		workerName := fmt.Sprintf("worker%d", i)
		nodes = append(nodes, Node{
			workerName,
			[]NodeComponent{
				NewResourceManager(workerName),
			},
		})
	}

	cluster.Nodes = nodes

	cluster.RegisterFunc(CTaskGen, shareTaskgenSetup, CommonTaskgenUpdate)
	cluster.RegisterFunc(CStateStorage, shareStateStorageSetup, shareStateStorageUpdate)
	cluster.RegisterFunc(CScheduler, shareSchedulerSetup, shareSchedulerUpdate)
	cluster.RegisterFunc(CResouceManger, shareResourceManagerSetup, CommonResourceManagerUpdate)

	return cluster
}

func shareStateStorageSetup(comp interface{})  {
	storage := comp.(*StateStorage)
	storage.StartTime=storage.Os.GetTime()
	for i := 0; i < int(Config.NodeNum); i++ {
		nodeinfo := &NodeInfo{Config.NodeCpu, Config.NodeMemory, 0, 0}
		storage.Workers["worker"+fmt.Sprint(i)+":"+string(CResouceManger)] = nodeinfo.Clone()
	}
}
func shareStateStorageUpdate(comp interface{}) {
	storage := comp.(*StateStorage)
	t := storage.Os.GetTime().Sub(storage.StartTime)
	if t.Milliseconds()%stateCopyUpdateMS == stateCopyUpdateMS-1{


	}
	

}

func shareTaskgenSetup(c interface{}) {
	taskgen := c.(*TaskGen)
	taskgen.StartTime = taskgen.Os.GetTime()
	for i := 0; i < ShareSchdulerNum; i++ {
		taskgen.Receivers = append(taskgen.Receivers,
			fmt.Sprintf("master%d", i)+":"+string(CScheduler),
		)
	}
}

func shareSchedulerSetup(s interface{}) {

}
func shareSchedulerUpdate(s interface{}) {

}
func shareResourceManagerSetup(r interface{}) {

}
