package main

import (
	"fmt"
	"simds-standalone/common"
)

// BuildCenterCluster 建立中心化调度的集群
// 中心化集群有三类实体 user1 任务发生器, master1 调度器 , Config.NodeNum 个worker
// 其三类实体分别有 Taskgen Schdueler ResourceManager 组件
// 最后注册这些组件的初始化和运行更新行为
func BuildCenterCluster() Cluster {
	var cluster = createCluster()
	var nodes []Node
	nodes = append(nodes, Node{
		"user1",
		[]NodeComponent{
			NewTaskGen("user1"),
		},
	})

	nodes = append(nodes, Node{
		"master1",
		[]NodeComponent{
			NewScheduler("master1"),
		},
	})
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
	cluster.RegisterFunc(CTaskGen, CenterTaskgenSetup, CommonTaskgenUpdate)
	cluster.RegisterFunc(CScheduler, CenterSchedulerSetup, CenterSchedulerUpdate)
	cluster.RegisterFunc(CResouceManger, CenterResourceManagerSetup, CommonResourceManagerUpdate)

	return cluster
}

// CenterTaskgenSetup 对中心化集群的任务发生器进行初始化
// 发送对象只有一个，发送时间未测试开始时
func CenterTaskgenSetup(c interface{}) {
	taskgen := c.(*TaskGen)
	taskgen.StartTime = taskgen.Os.GetTime()
	taskgen.Receivers = append(taskgen.Receivers, "master1"+":"+string(CScheduler))
}

// CenterSchedulerSetup 模拟开始时对中心化集群的调度器组件进行初始化
// 调度器注册所有的worker信息，包括任务发送地址和节点资源信息
func CenterSchedulerSetup(comp interface{}) {
	scheduler := comp.(*Scheduler)
	for i := 0; i < int(Config.NodeNum); i++ {

		nodeAddr := "worker" + fmt.Sprint(i) + ":" + string(CResouceManger)
		nodeinfo := &NodeInfo{nodeAddr, Config.NodeCpu, Config.NodeMemory, 0, 0}
		scheduler.Workers["worker"+fmt.Sprint(i)+":"+string(CResouceManger)] = nodeinfo.Clone()
	}

}

// CenterSchedulerUpdate 模拟器每次tick时对中心化集群的调度器组件进行初始化
// 从网络中取出消息进行处理，然后进行有次数限制的调度动作
func CenterSchedulerUpdate(comp interface{}) {

	scheduler := comp.(*Scheduler)

	for !scheduler.Os.Net().Empty() {
		newMessage, err := scheduler.Os.Net().Recv()
		if err != nil {
			panic(err)
		}

		if newMessage.Content == "TaskDispense" {
			task := newMessage.Body.(TaskInfo)
			task.Status = "WaitSchedule"
			scheduler.WaitSchedule.InQueue(task)
			LogInfo(scheduler.Os, "received TaskDispense", task)
		}

		if newMessage.Content == "TaskFinish" {
			taskInfo := newMessage.Body.(TaskInfo)
			scheduler.Workers[newMessage.From].SubAllocated(taskInfo.CpuRequest, taskInfo.MemoryRequest)
			LogInfo(scheduler.Os, "received TaskFinish", newMessage.From, taskInfo)
		}

	}

	var maxSchedulerTime = int(Config.SchedulerPerformance)
	for i := 0; i < maxSchedulerTime; i++ {

		task, err := scheduler.WaitSchedule.Dequeue()
		if err != nil {
			break
		}

		dstWorker, ok := schdulingAlgorithm(scheduler, &task)
		if ok {
			task.Worker = dstWorker
			task.Status = "Allocated"
			scheduler.Workers[task.Worker].AddAllocated(task.CpuRequest, task.MemoryRequest)
			newMessage := Message{
				From:    scheduler.Os.Net().GetAddr(),
				To:      task.Worker,
				Content: "TaskRun",
				Body:    task,
			}
			scheduler.Os.Net().Send(newMessage)
			LogInfo(scheduler.Os, "sendtask to", task.Worker, task)
		} else {
			scheduler.WaitSchedule.InQueueFront(task)

		}

	}
}

// schdulingAlgorithm 简单的调度算法，找到第一个合适调度的节点,找不到 ok返回false
func schdulingAlgorithm(scheduler *Scheduler, task *TaskInfo) (dstAddr string, ok bool) {
	dstAddr = ""

	keys := make([]string, 0, len(scheduler.Workers))
	for k := range scheduler.Workers {
		keys = append(keys, k)
	}
	common.ShuffleStringSlice(keys)
	for _, workerAddr := range keys {
		nodeinfo := scheduler.Workers[workerAddr]
		if nodeinfo.CanAllocate(task.CpuRequest, task.MemoryRequest) {
			dstAddr = workerAddr
		}
	}

	if dstAddr == "" {
		return dstAddr, false
	}
	return dstAddr, true
}

// CenterResourceManagerSetup 资源管理初始化，所有节点会发送任务结束通知给master1
func CenterResourceManagerSetup(comp interface{}) {
	rm := comp.(*ResourceManager)
	rm.TaskFinishReceiver = "master1" + ":" + string(CScheduler)

}
