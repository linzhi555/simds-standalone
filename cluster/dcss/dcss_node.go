package dcss

import (
	"math/rand"
	"os/exec"
	"time"

	"simds-standalone/cluster/base"
	"simds-standalone/cluster/lib"
	"simds-standalone/config"
)

// Scheduler 组件
type DcssNode struct {
	base.BasicActor
	TaskMap     map[string]*lib.TaskInfo
	RunningTask map[string]*lib.TaskInfo
	Neighbors   map[string]*lib.NodeInfo
	LocalNode   *lib.NodeInfo
}

func (node *DcssNode) Debug() {

}

func NewDcssNode(nodeName string) *DcssNode {
	newNode := DcssNode{
		BasicActor: base.BasicActor{
			Host: nodeName,
		},
		Neighbors: make(map[string]*lib.NodeInfo),
		LocalNode: &lib.NodeInfo{
			Addr:   nodeName,
			Cpu:    config.Val.NodeCpu,
			Memory: config.Val.NodeMemory,
		},
	}
	return &newNode

}

// DcssNodeUpdate 模拟器每次tick时对分布式集群的调度器组件进行初始化
// 调度器组件可以自己收到任务直接运行，也可以将任务进行转发，之后处理转发失败以及成功信
// 息，同时也要处理其他同类节点转发请求
func (node *DcssNode) Update(msg base.Message) {

	switch msg.Head {
	case "SignalBoot":
		time.Sleep(time.Millisecond * 1)
	case "TaskDispense":
		node.dcssTaskDispenseHandle(msg)
	case "TaskDivide":
		node.dcssTaskDivideHandle(msg)
	case "TaskDivideConfirm":
		node.dcssTaskDivideConfirmHandle(msg)
	case "TaskDivideAllocate":
		node.dcssTaskDivideAllocateHandle(msg)
	case "TaskDivideCancel":
		node.dcssTaskDivideCancelHandle(msg)
	case "TaskDivideReject":
		node.dcssTaskDivideRejectHandle(msg)
	case "TaskFinish":
		node.dcssFinishHandle(msg)
	}

}

func (node *DcssNode) dcssTaskDispenseHandle(newMessage base.Message) {
	task := newMessage.Body.(lib.TaskInfo)
	if node.LocalNode.CanAllocateTask(&task) {
		node.LocalNode.AddAllocated(task.CpuRequest, task.MemoryRequest)
		node._runTask(task)
		node.Os.Send(base.Message{
			From: node.Host,
			To:   task.User,
			Head: "TaskStart",
			Body: task,
		})
	} else {
		switch config.Val.DcssDividePolicy {
		case "always":
			node._dcssDivideTask(task)
		case "random":
			n := rand.Float32()
			if n < 0.3 {
				node._delaySchedule(task)
			} else if n < 0.6 {
				node._dcssDivideTask(task)
			} else {
				node._dispenseTask(task)
			}
		default:
			panic("wrong divide policy")

		}
	}
}
func (node *DcssNode) _delaySchedule(task lib.TaskInfo) {
	task.Status = "delaySchedule"
	task.LeftTime = time.Millisecond * 10
	node.RunningTask[task.Id] = &task
	node.Os.Run(func() {
		time.Sleep(time.Millisecond * 10)
		newMessage := base.Message{
			From: node.GetHostName(),
			To:   node.GetHostName(),
			Head: "TaskDispense",
			Body: task,
		}
		err := node.Os.Send(newMessage)
		if err != nil {
			panic(err)
		}
	})
}

func (node *DcssNode) _dcssDivideTask(task lib.TaskInfo) {
	task.Status = "DivideStage1"
	task.ScheduleFailCount = 0 // this is for count how many neibor reject this task
	keys := make([]string, 0, len(node.Neighbors))
	for k := range node.Neighbors {
		keys = append(keys, k)
	}

	for _, neibor := range keys {
		newMessage := base.Message{
			From: node.GetHostName(),
			To:   neibor,
			Head: "TaskDivide",
			Body: task,
		}
		err := node.Os.Send(newMessage)
		if err != nil {
			panic(err)
		}
	}
	task.Status = "DivideStage2"
	node.TaskMap[task.Id] = &task
}

func (node *DcssNode) _dispenseTask(task lib.TaskInfo) {
	keys := make([]string, 0, len(node.Neighbors))
	for k := range node.Neighbors {
		keys = append(keys, k)
	}
	neibors := keys
	dstNeibor := neibors[rand.Intn(len(neibors))]
	newMessage := base.Message{
		From: node.GetHostName(),
		To:   dstNeibor,
		Head: "TaskDispense",
		Body: task,
	}
	err := node.Os.Send(newMessage)
	if err != nil {
		panic(err)
	}
}

func (node *DcssNode) dcssTaskDivideHandle(newMessage base.Message) {
	task := newMessage.Body.(lib.TaskInfo)
	messageReply := newMessage
	messageReply.To = newMessage.From
	messageReply.From = newMessage.To
	if node.LocalNode.CanAllocateTask(&task) {
		node.LocalNode.AddAllocated(task.CpuRequest, task.MemoryRequest)
		messageReply.Head = "TaskDivideConfirm"
		task.Status = "needStart"
		node.TaskMap[task.Id] = &task
	} else {
		messageReply.Head = "TaskDivideReject"
	}
	err := node.Os.Send(messageReply)
	if err != nil {
		panic(err)
	}
}

func (node *DcssNode) dcssTaskDivideConfirmHandle(newMessage base.Message) {
	task := newMessage.Body.(lib.TaskInfo)
	t := node.TaskMap[task.Id]

	if t.Status == "DivideStage2" {
		t.Status = "DivideStage3"
		err := node.Os.Send(base.Message{
			From: newMessage.To,
			To:   newMessage.From,
			Head: "TaskDivideAllocate",
			Body: *node.TaskMap[task.Id],
		})
		if err != nil {
			panic(err)
		}
	} else if t.Status == "DivideStage3" {
		err := node.Os.Send(base.Message{
			From: newMessage.To,
			To:   newMessage.From,
			Head: "TaskDivideCancel",
			Body: *node.TaskMap[task.Id],
		})
		if err != nil {
			panic(err)
		}

	}
}
func (node *DcssNode) dcssTaskDivideAllocateHandle(newMessage base.Message) {
	task := newMessage.Body.(lib.TaskInfo)
	if t, ok := node.TaskMap[task.Id]; ok {
		if t.Status == "needStart" {
			node._runTask(*t)
			delete(node.TaskMap, t.Id)

			node.Os.Send(base.Message{
				From: node.Host,
				To:   task.User,
				Head: "TaskStart",
				Body: task,
			})

		}
	}
}

func (node *DcssNode) _runTask(t lib.TaskInfo) {
	t.StartTime = node.Os.GetTime()
	t.LeftTime = t.LifeTime
	t.Status = "start"
	node.RunningTask[t.Id] = &t
	node.Os.Run(func() {
		cmd := exec.Command("bash", "-c", t.Cmd)
		err := cmd.Run()
		if err != nil {
			panic(err)
		}
		newMessage := base.Message{
			From: node.GetHostName(),
			To:   node.GetHostName(),
			Head: "TaskFinish",
			Body: t,
		}
		err = node.Os.Send(newMessage)
		if err != nil {
			panic(err)
		}

	})
}

func (node *DcssNode) dcssTaskDivideCancelHandle(newMessage base.Message) {
	task := newMessage.Body.(lib.TaskInfo)

	if t, ok := node.TaskMap[task.Id]; ok {
		if t.Status == "needStart" {
			node.LocalNode.SubAllocated(task.CpuRequest, task.MemoryRequest)
			delete(node.TaskMap, task.Id)
		}
	}
}

func (node *DcssNode) dcssTaskDivideRejectHandle(newMessage base.Message) {
	task := newMessage.Body.(lib.TaskInfo)
	node.TaskMap[task.Id].ScheduleFailCount++
	neiborNum := config.Val.DcssNeibor
	// if all neibors reject this task, so we i have to dispense the task to a random neibors,
	// the distination neibors  may have a valid neibor to execute this task
	if node.TaskMap[task.Id].ScheduleFailCount == int32(neiborNum) {
		var taskCopy lib.TaskInfo = *(node.TaskMap[task.Id])
		node._dispenseTask(taskCopy)
		delete(node.TaskMap, taskCopy.Id)
	}
}

func (node *DcssNode) dcssFinishHandle(newMessage base.Message) {
	task := newMessage.Body.(lib.TaskInfo)
	task.Status = "finish"
	node.LocalNode.SubAllocated(task.CpuRequest, task.MemoryRequest)

	node.Os.Send(base.Message{
		From: node.Host,
		To:   task.User,
		Head: "TaskFinish",
		Body: task,
	})

	delete(node.RunningTask, task.Id)

}
func (node *DcssNode) SimulateTasksUpdate() {
	for _, t := range node.RunningTask {
		if t.LeftTime > 0 {
			t.LeftTime -= (time.Second / time.Duration(config.Val.FPS))
			messageType := ""
			if t.Status == "start" {
				messageType = "TaskFinish"
			} else if t.Status == "delaySchedule" {
				messageType = "TaskDispense"
			}
			if t.LeftTime <= 0 {
				newMessage := base.Message{
					From: node.GetHostName(),
					To:   node.GetHostName(),
					Head: messageType,
					Body: *t,
				}
				err := node.Os.Send(newMessage)
				if err != nil {
					panic(err)
				}

			}
		}
	}
}
