package main

import (
	"fmt"
	"math/rand"
	"simds-standalone/config"
	"strconv"
	"strings"
)

// BuildDCSSCluster 建立分布式调度的集群
// 中心化集群有三类实体 user1 任务发生器, config.Val.NodeNum 个 node 实体
// user1 有 Taskgen组件， node实体有既有 Scheduler 也有 ResourceManager 组件（既是调度器也能worker）
func BuildDCSSCluster() Cluster {
	var cluster = createCluster()
	var nodes []Node
	nodes = append(nodes, Node{
		"user1",
		[]NodeComponent{
			NewTaskGen("user1"),
		},
	})

	for i := 0; i < int(config.Val.NodeNum); i++ {

		nodeName := fmt.Sprintf("node%d", i)
		nodes = append(nodes, Node{
			nodeName,
			[]NodeComponent{
				NewScheduler(nodeName),
				NewResourceManager(nodeName),
			},
		})

	}
	cluster.Nodes = nodes
	cluster.RegisterFunc(CTaskGen, DcssTaskgenSetup, CommonTaskgenUpdate)
	cluster.RegisterFunc(CScheduler, DcssSchedulerSetup, DcssSchedulerUpdate)
	cluster.RegisterFunc(CResouceManger, DcssResourceManagerSetup, CommonResourceManagerUpdate)

	return cluster
}

// DcssTaskgenSetup 对中心化集群的任务发生器进行初始化
// 发送对象是集群的所有node类实体的Scheduler组件
func DcssTaskgenSetup(c interface{}) {
	taskgen := c.(*TaskGen)
	taskgen.StartTime = taskgen.Os.GetTime()
	for i := 0; i < int(config.Val.NodeNum); i++ {

		taskgen.Receivers = append(taskgen.Receivers,
			"node"+fmt.Sprint(i)+":"+string(CScheduler),
		)
	}

}

// DcssSchedulerSetup 模拟开始时对分布式集群调度器组件进行初始化
// 和中心化调度器的不同 workers 存储的时邻域信息
func DcssSchedulerSetup(comp interface{}) {
	scheduler := comp.(*Scheduler)

	// init local node info
	scheduler.LocalNode = &NodeInfo{scheduler.Os.Net().GetAddr(), config.Val.NodeCpu, config.Val.NodeMemory, 0, 0}

	// init neibors
	neiborNum := int(config.Val.DcssNeibor)
	allNodeNum := int(config.Val.NodeNum)
	neiborRandom := int(config.Val.DcssNeiborRandomP * (float32(config.Val.DcssNeibor)))

	var neibors []string = make([]string, 0, neiborNum)

	selfIndex, err := strconv.Atoi(strings.TrimLeft(scheduler.Host, "node"))
	if err != nil {
		panic(err)
	}

	for _, neiborIndex := range getNeigbor(allNodeNum, selfIndex, neiborNum, neiborRandom) {
		newNeibor := fmt.Sprintf("node%d:Scheduler", neiborIndex)
		neibors = append(neibors, newNeibor)
	}

	for _, n := range neibors {
		nodeInfo := &NodeInfo{n, config.Val.NodeCpu, config.Val.NodeMemory, 0, 0}
		scheduler.Workers[n] = nodeInfo.Clone()
	}

	// for debug
	keys := make([]string, 0, len(scheduler.Workers))
	for k := range scheduler.Workers {
		keys = append(keys, k)
	}
	LogInfo(scheduler.Os, fmt.Sprintf("index %d,Now,I have %d neibor, they are %s", selfIndex, len(scheduler.Workers), keys))

}

// 创建邻域的算法，输入 一个节点的编号（selfIndex ） 返回其领域的一系列编号
func getNeigbor(allNodes int, selfIndex int, neiborNum int, neiborRandom int) []int {
	if neiborNum <= 2 {
		panic("neigborNum can not smaller than 2 ")
	}

	if allNodes < neiborNum*2 {
		panic("allNodes num can not smaller than neigborNum*2")
	}

	if selfIndex < 0 || selfIndex > allNodes-1 {
		panic("wrong index")
	}

	if neiborRandom > neiborNum {
		panic("to many random neigbor")
	}

	neinodes := make([]int, 0)
	res := make(map[int]struct{})
	// 规则邻居的编号
	leftNei := 0
	rightNei := 0
	for q := 0; q < neiborNum-neiborRandom; q++ {
		if q%2 == 0 {
			//加左邻居
			nextLeft := (selfIndex - leftNei - 1 + allNodes) % allNodes
			res[nextLeft] = struct{}{}
			leftNei++
		} else {
			nextRight := (selfIndex + rightNei + 1 + allNodes) % allNodes
			res[nextRight] = struct{}{}
			rightNei++
		}
	}
	// 随机邻居的编号
	for len(res) < neiborNum {
		index := rand.Intn(allNodes)
		if index == selfIndex {
			continue
		}
		res[index] = struct{}{}
	}
	for k := range res {
		neinodes = append(neinodes, k)
	}
	return neinodes
}

// DcssSchedulerUpdate 模拟器每次tick时对分布式集群的调度器组件进行初始化
// 调度器组件可以自己收到任务直接运行，也可以将任务进行转发，之后处理转发失败以及成功信
// 息，同时也要处理同类Scheduler的转发请求
func DcssSchedulerUpdate(comp interface{}) {
	scheduler := comp.(*Scheduler)

	for !scheduler.Os.Net().Empty() {
		newMessage, err := scheduler.Os.Net().Recv()
		if err != nil {
			panic(err)
		}
		LogInfo(scheduler.Os, scheduler.Os.Net().GetAddr(), "received", newMessage.Content, newMessage.Body)
		switch newMessage.Content {
		case "TaskDispense":
			dcssTaskDispenseHandle(scheduler, newMessage)
		case "TaskDivide":
			dcssTaskDivideHandle(scheduler, newMessage)
		case "TaskDivideConfirm":
			dcssTaskDivideConfirmHandle(scheduler, newMessage)
		case "TaskDivideAllocate":
			dcssTaskDivideAllocateHandle(scheduler, newMessage)
		case "TaskDivideCancel":
			dcssTaskDivideCancelHandle(scheduler, newMessage)
		case "TaskDivideReject":
			dcssTaskDivideRejectHandle(scheduler, newMessage)
		case "TaskFinish":
			dcssFinishHandle(scheduler, newMessage)
		}

	}

	// 在dcss 调度器当中 待调度队列的任务会在自身资源空闲时运行
	if !scheduler.WaitSchedule.Empty() {
		// 平均 每 10 次更新(10*0.1 = 1ms)调度一次
		if rand.Float32() < 0.1 {
			scheduleTask(scheduler)
		}
	}

}

func scheduleTask(scheduler *Scheduler) {
	task, err := scheduler.WaitSchedule.Dequeue()
	if err != nil {
		panic(err)
	}
	if scheduler.LocalNode.CanAllocateTask(&task) {
		_runLocally(scheduler, task)
	} else {
		switch config.Val.DcssDividePolicy {
		case "always":
			LogInfo(scheduler.Os, "lack of resource, divide process start", task)
			_dcssDivideTask(scheduler, task)
		case "random":
			n := rand.Float32()
			if n < 0.1 {
				LogInfo(scheduler.Os, "lack of resource, dispense task directly", task)
				_dispenseTask(scheduler, task)
			} else if n < 0.2 {
				LogInfo(scheduler.Os, "lack of resource, divide process start", task)
				_dcssDivideTask(scheduler, task)
			} else {
				LogInfo(scheduler.Os, "lack of resource, run task later", task, scheduler.WaitSchedule.Len())
				scheduler.WaitSchedule.InQueueFront(task)
			}
		default:
			panic("wrong divide policy")

		}
	}
}

func dcssTaskDispenseHandle(scheduler *Scheduler, newMessage Message) {
	task := newMessage.Body.(TaskInfo)
	task.Status = "Scheduling"
	scheduler.WaitSchedule.InQueue(task)

}

func _runLocally(scheduler *Scheduler, task TaskInfo) {
	task.Status = "allocate"
	scheduler.LocalNode.AddAllocated(task.CpuRequest, task.MemoryRequest)
	_dcssChangeTaskStatusLocally(scheduler, &task, "TaskRun")
}

func _dcssDivideTask(scheduler *Scheduler, task TaskInfo) {
	LogInfo(scheduler.Os, "start divide ", task)
	task.Status = "DivideStage1"
	task.ScheduleFailCount = 0 // this is for count how many neibor reject this task
	keys := make([]string, 0, len(scheduler.Workers))
	for k := range scheduler.Workers {
		keys = append(keys, k)
	}

	for _, neibor := range keys {
		newMessage := Message{
			From:    scheduler.Os.Net().GetAddr(),
			To:      neibor,
			Content: "TaskDivide",
			Body:    task,
		}
		err := scheduler.Os.Net().Send(newMessage)
		if err != nil {
			panic(err)
		}
	}
	task.Status = "DivideStage2"
	scheduler.TasksStatus[task.Id] = &task
}

func _dispenseTask(scheduler *Scheduler, task TaskInfo) {
	neibors := scheduler.GetAllWokersName()
	dstNeibor := neibors[rand.Intn(len(neibors))]
	newMessage := Message{
		From:    scheduler.Os.Net().GetAddr(),
		To:      dstNeibor,
		Content: "TaskDispense",
		Body:    task,
	}
	err := scheduler.Os.Net().Send(newMessage)
	if err != nil {
		panic(err)
	}
}

func dcssTaskDivideHandle(scheduler *Scheduler, newMessage Message) {
	task := newMessage.Body.(TaskInfo)
	messageReply := newMessage
	messageReply.To = newMessage.From
	messageReply.From = newMessage.To
	if scheduler.LocalNode.CanAllocateTask(&task) && scheduler.WaitSchedule.Empty() {
		messageReply.Content = "TaskDivideConfirm"
		task.Status = "needStart"
		scheduler.LocalNode.AddAllocated(task.CpuRequest, task.MemoryRequest)
		_dcssChangeTaskStatusLocally(scheduler, &task, "TaskPreAllocate")
		scheduler.TasksStatus[task.Id] = &task
	} else {
		messageReply.Content = "TaskDivideReject"
	}
	err := scheduler.Os.Net().Send(messageReply)
	if err != nil {
		panic(err)
	}
}
func dcssTaskDivideConfirmHandle(scheduler *Scheduler, newMessage Message) {
	task := newMessage.Body.(TaskInfo)
	t := scheduler.TasksStatus[task.Id]

	if t.Status == "DivideStage2" {
		t.Status = "DivideStage3"
		err := scheduler.Os.Net().Send(Message{
			From:    newMessage.To,
			To:      newMessage.From,
			Content: "TaskDivideAllocate",
			Body:    *scheduler.TasksStatus[task.Id],
		})
		if err != nil {
			panic(err)
		}

	} else if t.Status == "DivideStage3" {
		err := scheduler.Os.Net().Send(Message{
			From:    newMessage.To,
			To:      newMessage.From,
			Content: "TaskDivideCancel",
			Body:    *scheduler.TasksStatus[task.Id],
		})
		if err != nil {
			panic(err)
		}

	}
}
func dcssTaskDivideAllocateHandle(scheduler *Scheduler, newMessage Message) {
	task := newMessage.Body.(TaskInfo)
	if scheduler.TasksStatus[task.Id].Status == "needStart" {
		_dcssChangeTaskStatusLocally(scheduler, &task, "TaskStart")
	}
}

func dcssTaskDivideCancelHandle(scheduler *Scheduler, newMessage Message) {
	task := newMessage.Body.(TaskInfo)
	if scheduler.TasksStatus[task.Id].Status == "needStart" {
		_dcssChangeTaskStatusLocally(scheduler, &task, "TaskCancelAllocate")
	}
}

func dcssTaskDivideRejectHandle(scheduler *Scheduler, newMessage Message) {
	task := newMessage.Body.(TaskInfo)
	scheduler.TasksStatus[task.Id].ScheduleFailCount++
	neiborNum := config.Val.DcssNeibor
	// if all neibors reject this task, so we i have to dispense the task to a random neibors,
	// the distination neibors  may have a valid neibor to execute this task
	if scheduler.TasksStatus[task.Id].ScheduleFailCount == int32(neiborNum) {
		var taskCopy TaskInfo = *(scheduler.TasksStatus[task.Id])
		_dispenseTask(scheduler, taskCopy)
		delete(scheduler.TasksStatus, taskCopy.Id)
		LogInfo(scheduler.Os, "TaskDivide finally fail, start a new TaskDispense", newMessage.Body)
	}
}

func _dcssChangeTaskStatusLocally(scheduler *Scheduler, task *TaskInfo, contentType string) {
	dstWorker := scheduler.Host + ":" + string(CResouceManger)
	newMessage := Message{
		From:    scheduler.Os.Net().GetAddr(),
		To:      dstWorker,
		Content: contentType,
		Body:    *task,
	}
	err := scheduler.Os.Net().Send(newMessage)
	if err != nil {
		panic(err)
	}
}

func dcssFinishHandle(scheduler *Scheduler, newMessage Message) {
	task := newMessage.Body.(TaskInfo)
	scheduler.LocalNode.SubAllocated(task.CpuRequest, task.MemoryRequest)

}

// DcssResourceManagerSetup 资源管理初始化，所有节点会发送任务结束通知给相同host的Scheduler组件
func DcssResourceManagerSetup(comp interface{}) {
	rm := comp.(*ResourceManager)
	rm.TaskFinishReceiver = rm.Host + ":" + string(CScheduler)
}
