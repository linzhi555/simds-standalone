package main

import (
	"fmt"
	"time"
)

const shareSchdulerNum = 3

//const stateCopyUpdateMS = 200

// BuildShareStateCluster 建立分布式调度的集群
// 中心化集群有四类实体 user globalStateStorage master worker
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

	for i := 0; i < shareSchdulerNum; i++ {
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
		nodeAddr := "worker" + fmt.Sprint(i) + ":" + string(CResouceManger)
		nodeinfo := &NodeInfo{nodeAddr, Config.NodeCpu, Config.NodeMemory, 0, 0}
		storage.Workers[nodeAddr] = nodeinfo
	}
}

type CommitReply struct {
	NodeInfo
	TaskInfo
}

func (CommitReply) MessageBody() {}

func shareStateStorageUpdate(comp interface{}) {
	storage := comp.(*StateStorage)
	timeNow := storage.Os.GetTime()

	for !storage.Os.Net().Empty() {
		newMessage, err := storage.Os.Net().Recv()
		if err != nil {
			panic(err)
		}
		switch newMessage.Content {
		case "TaskCommit":
			task := newMessage.Body.(TaskInfo)
			if storage.Workers[task.Worker].CanAllocate(task.CpuRequest, task.MemoryRequest) {
				task.Status = "CommitSuccess"
				storage.Workers[task.Worker].AddAllocated(task.CpuRequest, task.MemoryRequest)
				err := storage.Os.Net().Send(Message{
					From:    storage.Os.Net().GetAddr(),
					To:      task.Worker,
					Content: "TaskRun",
					Body:    task,
				})
				if err != nil {
					panic(err)
				}

				LogInfo(storage.Os, "task commit success", task.Worker, task)
			} else {
				task.Status = "CommitFail"
				err := storage.Os.Net().Send(Message{
					From:    storage.Os.Net().GetAddr(),
					To:      newMessage.From,
					Content: "TaskCommitFail",
					Body: CommitReply{
						*storage.Workers[task.Worker],
						task,
					},
				})
				if err != nil {
					panic(err)
				}

				LogInfo(storage.Os, "task commit fail ", storage.Workers[task.Worker], task)
			}

		case "TaskFinish":
			taskInfo := newMessage.Body.(TaskInfo)
			storage.Workers[newMessage.From].SubAllocated(taskInfo.CpuRequest, taskInfo.MemoryRequest)
			LogInfo(storage.Os, "received TaskFinish", newMessage.From, taskInfo)

		default:
			panic("wrong type message,please check who has send this message to me!")
		}
	}

	if timeNow.Sub(storage.LastSendTime).Milliseconds() > int64(Config.StateUpdatePeriod) {
		storage.LastSendTime = timeNow

		stateCopy := storage.StateCopy()
		for i := 0; i < shareSchdulerNum; i++ {
			dist := fmt.Sprintf("master%d", i) + ":" + string(CScheduler)
			err := storage.Os.Net().Send(Message{
				From:    storage.Os.Net().GetAddr(),
				To:      dist,
				Content: "ClusterStateCopy",
				Body:    *stateCopy.Clone(),
			})
			if err != nil {
				panic(err)
			}

		}

	}
}

func shareTaskgenSetup(c interface{}) {
	taskgen := c.(*TaskGen)
	//wait the schedulers until state updated, then launch the taskgen
	taskgen.StartTime = taskgen.Os.GetTime().Add(time.Millisecond * time.Duration(Config.StateUpdatePeriod) * 2)
	for i := 0; i < shareSchdulerNum; i++ {
		taskgen.Receivers = append(taskgen.Receivers,
			fmt.Sprintf("master%d", i)+":"+string(CScheduler),
		)
	}
}

func shareSchedulerSetup(_ interface{}) {

}
func shareSchedulerUpdate(comp interface{}) {
	scheduler := comp.(*Scheduler)

	for !scheduler.Os.Net().Empty() {
		newMessage, err := scheduler.Os.Net().Recv()
		if err != nil {
			panic(err)
		}
		switch newMessage.Content {
		case "TaskDispense":
			task := newMessage.Body.(TaskInfo)
			task.Status = "WaitSchedule"
			scheduler.WaitSchedule.InQueue(task)
			LogInfo(scheduler.Os, "received TaskDispense", task)
		case "TaskCommitFail":
			reply := newMessage.Body.(CommitReply)
			task := reply.TaskInfo
			task.Status = "WaitSchedule"
			scheduler.WaitSchedule.InQueueFront(task)

			nodeinfo := reply.NodeInfo
			scheduler.Workers[task.Worker] = nodeinfo.Clone()

			LogInfo(scheduler.Os, "reschedule task", task)
		case "ClusterStateCopy":
			nodeinfoList := newMessage.Body.(Vec[NodeInfo])
			//for k := range scheduler.Workers {
			//	delete(scheduler.Workers, k)
			//}
			for _, ni := range nodeinfoList {
				scheduler.Workers[ni.Addr] = ni.Clone()
			}
		}
	}

	var maxScheduleTimes = schdulingAlgorithmTimes(Config.SchedulerPerformance)
	for i := 0; i < maxScheduleTimes; i++ {
		task, err := scheduler.WaitSchedule.Dequeue()
		if err != nil {
			break
		}
		dstWorker, ok := schdulingAlgorithm(scheduler, &task)
		if ok {
			task.Worker = dstWorker
			task.Status = "TryAllocate"
			scheduler.Workers[task.Worker].AddAllocated(task.CpuRequest, task.MemoryRequest)
			newMessage := Message{
				From:    scheduler.Os.Net().GetAddr(),
				To:      fmt.Sprintf("globalStateStorage:%s", string(CStateStorage)),
				Content: "TaskCommit",
				Body:    task,
			}
			err := scheduler.Os.Net().Send(newMessage)
			if err != nil {
				panic(err)
			}
			LogInfo(scheduler.Os, "try to commit task allocate to globalStateStorage", task.Worker, task)
		} else {
			scheduler.WaitSchedule.InQueueFront(task)
			break
		}
	}
}
func shareResourceManagerSetup(comp interface{}) {
	rm := comp.(*ResourceManager)
	rm.TaskFinishReceiver = "globalStateStorage" + ":" + string(CStateStorage)
}
