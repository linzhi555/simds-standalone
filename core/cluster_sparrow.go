package core

// import (
// 	"fmt"
// 	"math/rand"
// 	"simds-standalone/config"
// 	"strings"
// )

// const SparrowProbeNum = 2

// func getSparrowSchedulerNum() int {
// 	return int(config.Val.SparrowSchedulerNumFactor * float32(config.Val.NodeNum))

// }

// // BuildSparrowCluster 建立分布式调度的集群
// // 中心化集群有三类实体 user1 任务发生器, config.Val.NodeNum 个 node 实体
// // user1 有 Taskgen组件， node实体有既有 Scheduler 也有 ResourceManager 组件（既是调度器也能worker）
// func BuildSparrowCluster() Cluster {
// 	var cluster = createCluster()
// 	var nodes []Node
// 	nodes = append(nodes, Node{
// 		"user1",
// 		[]Component{
// 			NewTaskGen("user1"),
// 		},
// 	})

// 	for i := 0; i < getSparrowSchedulerNum(); i++ {

// 		nodeName := fmt.Sprintf("scheduler%d", i)
// 		nodes = append(nodes, Node{
// 			nodeName,
// 			[]Component{
// 				NewScheduler(nodeName),
// 			},
// 		})

// 	}

// 	for i := 0; i < int(config.Val.NodeNum); i++ {

// 		nodeName := fmt.Sprintf("worker%d", i)
// 		nodes = append(nodes, Node{
// 			nodeName,
// 			[]Component{
// 				NewScheduler(nodeName),
// 				NewResourceManager(nodeName),
// 			},
// 		})
// 	}

// 	cluster.Nodes = nodes
// 	cluster.RegisterFunc(CTaskGen, SparrowTaskgenSetup, CommonTaskgenUpdate)
// 	cluster.RegisterFunc(CScheduler, SparrowSchedulerSetup, SparrowSchedulerUpdate)
// 	cluster.RegisterFunc(CResouceManger, SparrowResourceManagerSetup, CommonResourceManagerUpdate)

// 	return cluster
// }

// // SparrowTaskgenSetup 对中心化集群的任务发生器进行初始化
// // 发送对象是集群的所有sparrow所有的Scheduler组件
// func SparrowTaskgenSetup(c Component) {
// 	taskgen := c.(*TaskGen)
// 	taskgen.StartTime = taskgen.Os.GetTime()
// 	for i := 0; i < getSparrowSchedulerNum(); i++ {

// 		taskgen.Receivers = append(taskgen.Receivers,
// 			fmt.Sprintf("scheduler%d", i)+":"+string(CScheduler),
// 		)
// 	}
// }

// // SparrowSchedulerSetup 模拟开始时对分布式集群调度器组件进行初始化
// // 和中心化调度器的不同 workers 存储的时邻域信息
// func SparrowSchedulerSetup(comp Component) {
// 	scheduler := comp.(*Scheduler)

// 	if strings.HasPrefix(scheduler.Host, "worker") {
// 		scheduler.LocalNode = &NodeInfo{scheduler.Os.Net().GetAddr(), config.Val.NodeCpu, config.Val.NodeMemory, 0, 0}
// 	}

// 	if strings.HasPrefix(scheduler.Host, "scheduler") {

// 		scheduler.Os.LogInfo(scheduler.Os, fmt.Sprintf("sparrow scheduler Num is %d", getSparrowSchedulerNum()))
// 		for i := 0; i < int(config.Val.NodeNum); i++ {
// 			nodeAddr := "worker" + fmt.Sprint(i) + ":" + string(CScheduler)
// 			nodeinfo := &NodeInfo{nodeAddr, config.Val.NodeCpu, config.Val.NodeMemory, 0, 0}
// 			scheduler.Workers["worker"+fmt.Sprint(i)+":"+string(CScheduler)] = nodeinfo.Clone()
// 		}
// 	}

// }

// // SparrowSchedulerUpdate 模拟器每次tick时对分布式集群的调度器组件进行初始化
// // 调度器组件可以自己收到任务直接运行，也可以将任务进行转发，之后处理转发失败以及成功信
// // 息，同时也要处理同类Scheduler的转发请求
// func SparrowSchedulerUpdate(comp Component) {
// 	scheduler := comp.(*Scheduler)

// 	for !scheduler.Os.Net().Empty() {
// 		newMessage, err := scheduler.Os.Net().Recv()
// 		if err != nil {
// 			panic(err)
// 		}
// 		scheduler.Os.LogInfo(scheduler.Os, "received_from", newMessage.From, newMessage.Content, newMessage.Body)
// 		switch newMessage.Content {
// 		case "TaskDispense":
// 			sparrowTaskDispenseHandle(scheduler, newMessage)
// 		case "TaskDivide":
// 			sparrowTaskDivideHandle(scheduler, newMessage)
// 		case "TaskDivideConfirm":
// 			sparrowTaskDivideConfirmHandle(scheduler, newMessage)
// 		case "TaskDivideAllocate":
// 			sparrowTaskDivideAllocateHandle(scheduler, newMessage)
// 		case "TaskDivideCancel":
// 			sparrowTaskDivideCancelHandle(scheduler, newMessage)
// 		case "TaskDivideReject":
// 			sparrowTaskDivideRejectHandle(scheduler, newMessage)
// 		case "TaskFinish":
// 			sparrowFinishHandle(scheduler, newMessage)
// 		}

// 	}

// 	// 在sparrow 调度器当中 待调度队列的任务会在自身资源空闲时运行
// 	if !scheduler.WaitSchedule.Empty() {
// 		// 平均 每 10 次更新(10*0.1 = 1ms)调度一次
// 		if rand.Float32() < 0.1 {
// 			_sparrowScheduleTask(scheduler)
// 		}
// 	}

// }

// func _sparrowScheduleTask(scheduler *Scheduler) {
// 	task, err := scheduler.WaitSchedule.Dequeue()
// 	if err != nil {
// 		panic(err)
// 	}
// 	_sparrowDivideTask(scheduler, task)
// }

// func sparrowTaskDispenseHandle(scheduler *Scheduler, newMessage Message) {
// 	task := newMessage.Body.(TaskInfo)
// 	task.Status = "Scheduling"
// 	scheduler.WaitSchedule.InQueue(task)

// }

// func _sparrowDivideTask(scheduler *Scheduler, task TaskInfo) {
// 	scheduler.Os.LogInfo(scheduler.Os, "start divide ", task)
// 	task.Status = "DivideStage1"
// 	task.ScheduleFailCount = 0 // this is for count how many neibor reject this task

// 	selected := map[int]bool{}
// 	for len(selected) < SparrowProbeNum {
// 		workerId := rand.Intn(len(scheduler.Workers))
// 		if _, exsited := selected[workerId]; exsited {
// 			continue
// 		}
// 		selected[workerId] = true
// 		workname := "worker" + fmt.Sprint(workerId) + ":" + string(CScheduler)
// 		newMessage := Message{
// 			From:    scheduler.Os.Net().GetAddr(),
// 			To:      workname,
// 			Content: "TaskDivide",
// 			Body:    task,
// 		}
// 		err := scheduler.Os.Net().Send(newMessage)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// 	task.Status = "DivideStage2"
// 	scheduler.TasksStatus[task.Id] = &task
// }

// func sparrowTaskDivideHandle(scheduler *Scheduler, newMessage Message) {
// 	task := newMessage.Body.(TaskInfo)
// 	messageReply := newMessage
// 	messageReply.To = newMessage.From
// 	messageReply.From = newMessage.To
// 	if scheduler.LocalNode.CanAllocateTask(&task) && scheduler.WaitSchedule.Empty() {
// 		messageReply.Content = "TaskDivideConfirm"
// 		task.Status = "needStart"
// 		scheduler.LocalNode.AddAllocated(task.CpuRequest, task.MemoryRequest)
// 		_sparrowChangeTaskStatusLocally(scheduler, &task, "TaskPreAllocate")
// 		scheduler.TasksStatus[task.Id] = &task
// 	} else {
// 		messageReply.Content = "TaskDivideReject"
// 	}
// 	err := scheduler.Os.Net().Send(messageReply)
// 	if err != nil {
// 		panic(err)
// 	}
// }
// func sparrowTaskDivideConfirmHandle(scheduler *Scheduler, newMessage Message) {
// 	task := newMessage.Body.(TaskInfo)
// 	t := scheduler.TasksStatus[task.Id]

// 	if t.Status == "DivideStage2" {
// 		t.Status = "DivideStage3"
// 		err := scheduler.Os.Net().Send(Message{
// 			From:    newMessage.To,
// 			To:      newMessage.From,
// 			Content: "TaskDivideAllocate",
// 			Body:    *scheduler.TasksStatus[task.Id],
// 		})
// 		if err != nil {
// 			panic(err)
// 		}

// 	} else if t.Status == "DivideStage3" {
// 		err := scheduler.Os.Net().Send(Message{
// 			From:    newMessage.To,
// 			To:      newMessage.From,
// 			Content: "TaskDivideCancel",
// 			Body:    *scheduler.TasksStatus[task.Id],
// 		})
// 		if err != nil {
// 			panic(err)
// 		}

// 		delete(scheduler.TasksStatus, task.Id)
// 	}
// }
// func sparrowTaskDivideAllocateHandle(scheduler *Scheduler, newMessage Message) {
// 	task := newMessage.Body.(TaskInfo)
// 	if scheduler.TasksStatus[task.Id].Status == "needStart" {
// 		_sparrowChangeTaskStatusLocally(scheduler, &task, "TaskStart")
// 	}
// }

// func sparrowTaskDivideCancelHandle(scheduler *Scheduler, newMessage Message) {
// 	task := newMessage.Body.(TaskInfo)
// 	if scheduler.TasksStatus[task.Id].Status == "needStart" {
// 		_sparrowChangeTaskStatusLocally(scheduler, &task, "TaskCancelAllocate")
// 	}
// }

// func sparrowTaskDivideRejectHandle(scheduler *Scheduler, newMessage Message) {
// 	task := newMessage.Body.(TaskInfo)
// 	scheduler.TasksStatus[task.Id].ScheduleFailCount++
// 	if scheduler.TasksStatus[task.Id].ScheduleFailCount == int32(SparrowProbeNum) {
// 		var taskCopy TaskInfo = *(scheduler.TasksStatus[task.Id])
// 		taskCopy.ScheduleFailCount = 0
// 		scheduler.WaitSchedule.InQueueFront(taskCopy)
// 		delete(scheduler.TasksStatus, task.Id)
// 		scheduler.Os.LogInfo(scheduler.Os, "Divide fail, push this task to queue again", newMessage.Body)
// 	}
// }

// func _sparrowChangeTaskStatusLocally(scheduler *Scheduler, task *TaskInfo, contentType string) {
// 	dstWorker := scheduler.Host + ":" + string(CResouceManger)
// 	newMessage := Message{
// 		From:    scheduler.Os.Net().GetAddr(),
// 		To:      dstWorker,
// 		Content: contentType,
// 		Body:    *task,
// 	}
// 	err := scheduler.Os.Net().Send(newMessage)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func sparrowFinishHandle(scheduler *Scheduler, newMessage Message) {
// 	task := newMessage.Body.(TaskInfo)
// 	scheduler.LocalNode.SubAllocated(task.CpuRequest, task.MemoryRequest)

// }

// // SparrowResourceManagerSetup 资源管理初始化，所有节点会发送任务结束通知给相同host的Scheduler组件
// func SparrowResourceManagerSetup(comp Component) {
// 	rm := comp.(*ResourceManager)
// 	rm.Node = NodeInfo{rm.Host, config.Val.NodeCpu, config.Val.NodeMemory, 0, 0}
// 	rm.TaskFinishReceiver = rm.Host + ":" + string(CScheduler)
// }
