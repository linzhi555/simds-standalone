package main

import "fmt"

const ShareSchdulerNum = 3
const stateCopyUpdateMS = 200

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

func shareStateStorageSetup(comp interface{}) {
	storage := comp.(*StateStorage)
	storage.LastSendTime = storage.Os.GetTime()
	for i := 0; i < int(Config.NodeNum); i++ {
		nodeinfo := &NodeInfo{Config.NodeCpu, Config.NodeMemory, 0, 0}
		storage.Workers["worker"+fmt.Sprint(i)+":"+string(CResouceManger)] = nodeinfo.Clone()
	}
}
func shareStateStorageUpdate(comp interface{}) {
	storage := comp.(*StateStorage)
	timeNow := storage.Os.GetTime()
	if timeNow.Sub(storage.LastSendTime).Milliseconds() > stateCopyUpdateMS {
		storage.LastSendTime = timeNow

		stateCopy := storage.StateCopy()
		for i := 0; i < ShareSchdulerNum; i++ {
			dist := fmt.Sprintf("master%d", i) + ":" + string(CScheduler)
			storage.Os.Net().Send(Message{
				From:    storage.Os.Net().GetAddr(),
				To:      dist,
				Content: "ClusterStateCopy",
				Body:    *stateCopy.Clone(),
			})
		}
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
func shareSchedulerUpdate(comp interface{}) {
	scheduler := comp.(*Scheduler)
	for !scheduler.Os.Net().Empty() {
		newMessage, err := scheduler.Os.Net().Recv()
		if err != nil {
			panic(err)
		}
		LogInfo(scheduler.Os, scheduler.Os.Net().GetAddr(), "received", newMessage.Content, newMessage.Body)
	}
}
func shareResourceManagerSetup(r interface{}) {

}
