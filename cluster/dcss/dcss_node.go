package dcss

import (
	"fmt"
	"math/rand"
	"os/exec"
	"simds-standalone/config"
	base "simds-standalone/core"
	"time"
)

// Scheduler 组件
type DcssNode struct {
	base.BasicNode
	TaskMap     map[string]*base.TaskInfo
	RunningTask map[string]*base.TaskInfo
	Neighbors   map[string]*base.NodeInfo
	LocalNode   *base.NodeInfo
}

func (node *DcssNode) Debug() {

}

// DcssNodeUpdate 模拟器每次tick时对分布式集群的调度器组件进行初始化
// 调度器组件可以自己收到任务直接运行，也可以将任务进行转发，之后处理转发失败以及成功信
// 息，同时也要处理其他同类节点转发请求
func (node *DcssNode) Update(msg base.Message) {

	switch msg.Content {
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
	task := newMessage.Body.(base.TaskInfo)
	if node.LocalNode.CanAllocateTask(&task) {
		node.LocalNode.AddAllocated(task.CpuRequest, task.MemoryRequest)
		node._runTask(task)
		node.Os.LogInfo("stdout", node.GetHostName(), "TaskRun", fmt.Sprint(task))
		node.Os.Send(base.Message{
			From:    node.Host,
			To:      task.User,
			Content: "TaskStart",
			Body:    task,
		})
	} else {
		switch config.Val.DcssDividePolicy {
		case "always":
			node.Os.LogInfo("stdout", node.GetHostName(), "lack of resource, divide process start", fmt.Sprint(task))
			node._dcssDivideTask(task)
		case "random":
			n := rand.Float32()
			if n < 0.3 {
				node.Os.LogInfo("stdout", node.GetHostName(), "schedule task later", fmt.Sprint(task))
				node._delaySchedule(task)
			} else if n < 0.6 {
				node.Os.LogInfo("stdout", node.GetHostName(), "divide process start", fmt.Sprint(task))
				node._dcssDivideTask(task)
			} else {
				node.Os.LogInfo("stdout", node.GetHostName(), "dispense task directly", fmt.Sprint(task))
				node._dispenseTask(task)
			}
		default:
			panic("wrong divide policy")

		}
	}
}
func (node *DcssNode) _delaySchedule(task base.TaskInfo) {
	task.Status = "delaySchedule"
	task.LeftTime = time.Millisecond * 10
	node.RunningTask[task.Id] = &task
	node.Os.Run(func() {
		time.Sleep(time.Millisecond * 10)
		newMessage := base.Message{
			From:    node.GetHostName(),
			To:      node.GetHostName(),
			Content: "TaskDispense",
			Body:    task,
		}
		err := node.Os.Send(newMessage)
		if err != nil {
			panic(err)
		}
	})
}

func (node *DcssNode) _dcssDivideTask(task base.TaskInfo) {
	node.Os.LogInfo("stdout", node.GetHostName(), "TaskDivide", fmt.Sprint(task))
	task.Status = "DivideStage1"
	task.ScheduleFailCount = 0 // this is for count how many neibor reject this task
	keys := make([]string, 0, len(node.Neighbors))
	for k := range node.Neighbors {
		keys = append(keys, k)
	}

	for _, neibor := range keys {
		newMessage := base.Message{
			From:    node.GetHostName(),
			To:      neibor,
			Content: "TaskDivide",
			Body:    task,
		}
		err := node.Os.Send(newMessage)
		if err != nil {
			panic(err)
		}
	}
	task.Status = "DivideStage2"
	node.TaskMap[task.Id] = &task
}

func (node *DcssNode) _dispenseTask(task base.TaskInfo) {
	keys := make([]string, 0, len(node.Neighbors))
	for k := range node.Neighbors {
		keys = append(keys, k)
	}
	neibors := keys
	dstNeibor := neibors[rand.Intn(len(neibors))]
	newMessage := base.Message{
		From:    node.GetHostName(),
		To:      dstNeibor,
		Content: "TaskDispense",
		Body:    task,
	}
	err := node.Os.Send(newMessage)
	if err != nil {
		panic(err)
	}
}

func (node *DcssNode) dcssTaskDivideHandle(newMessage base.Message) {
	task := newMessage.Body.(base.TaskInfo)
	messageReply := newMessage
	messageReply.To = newMessage.From
	messageReply.From = newMessage.To
	if node.LocalNode.CanAllocateTask(&task) {
		node.LocalNode.AddAllocated(task.CpuRequest, task.MemoryRequest)
		messageReply.Content = "TaskDivideConfirm"
		task.Status = "needStart"
		node.TaskMap[task.Id] = &task
		node.Os.LogInfo("stdout", node.GetHostName(), "TaskDivideConfirm", fmt.Sprint(task))
	} else {
		fmt.Print(node.LocalNode)
		messageReply.Content = "TaskDivideReject"
		node.Os.LogInfo("stdout", node.GetHostName(), "TaskDivideReject", fmt.Sprint(task))
	}
	err := node.Os.Send(messageReply)
	if err != nil {
		panic(err)
	}
}

func (node *DcssNode) dcssTaskDivideConfirmHandle(newMessage base.Message) {
	task := newMessage.Body.(base.TaskInfo)
	t := node.TaskMap[task.Id]

	if t.Status == "DivideStage2" {
		t.Status = "DivideStage3"
		err := node.Os.Send(base.Message{
			From:    newMessage.To,
			To:      newMessage.From,
			Content: "TaskDivideAllocate",
			Body:    *node.TaskMap[task.Id],
		})
		if err != nil {
			panic(err)
		}
		node.Os.LogInfo("stdout", node.GetHostName(), "TaskDivideAllocate", fmt.Sprint(t))
	} else if t.Status == "DivideStage3" {
		err := node.Os.Send(base.Message{
			From:    newMessage.To,
			To:      newMessage.From,
			Content: "TaskDivideCancel",
			Body:    *node.TaskMap[task.Id],
		})
		if err != nil {
			panic(err)
		}
		node.Os.LogInfo("stdout", node.GetHostName(), "TaskDivideCancel", fmt.Sprint(t))

	}
}
func (node *DcssNode) dcssTaskDivideAllocateHandle(newMessage base.Message) {
	task := newMessage.Body.(base.TaskInfo)
	if t, ok := node.TaskMap[task.Id]; ok {
		if t.Status == "needStart" {
			node._runTask(*t)
			delete(node.TaskMap, t.Id)
			node.Os.LogInfo("stdout", node.GetHostName(), "TaskRun", fmt.Sprint(*t))

			node.Os.Send(base.Message{
				From:    node.Host,
				To:      task.User,
				Content: "TaskStart",
				Body:    task,
			})

		}
	}
}

func (node *DcssNode) _runTask(t base.TaskInfo) {
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
			From:    node.GetHostName(),
			To:      node.GetHostName(),
			Content: "TaskFinish",
			Body:    t,
		}
		err = node.Os.Send(newMessage)
		if err != nil {
			panic(err)
		}

	})
}

func (node *DcssNode) dcssTaskDivideCancelHandle(newMessage base.Message) {
	task := newMessage.Body.(base.TaskInfo)

	if t, ok := node.TaskMap[task.Id]; ok {
		if t.Status == "needStart" {
			node.LocalNode.SubAllocated(task.CpuRequest, task.MemoryRequest)
			node.Os.LogInfo("stdout", node.GetHostName(), "TaskDivideCancel", fmt.Sprint(t))
			delete(node.TaskMap, task.Id)
		}
	}
}

func (node *DcssNode) dcssTaskDivideRejectHandle(newMessage base.Message) {
	task := newMessage.Body.(base.TaskInfo)
	node.TaskMap[task.Id].ScheduleFailCount++
	neiborNum := config.Val.DcssNeibor
	// if all neibors reject this task, so we i have to dispense the task to a random neibors,
	// the distination neibors  may have a valid neibor to execute this task
	if node.TaskMap[task.Id].ScheduleFailCount == int32(neiborNum) {
		var taskCopy base.TaskInfo = *(node.TaskMap[task.Id])
		node._dispenseTask(taskCopy)
		delete(node.TaskMap, taskCopy.Id)
		node.Os.LogInfo("stdout", "TaskDivide finally fail, start a new TaskDispense", fmt.Sprint(newMessage.Body))
	}
}

func (node *DcssNode) dcssFinishHandle(newMessage base.Message) {
	task := newMessage.Body.(base.TaskInfo)
	task.Status = "finish"
	node.LocalNode.SubAllocated(task.CpuRequest, task.MemoryRequest)

	node.Os.Send(base.Message{
		From:    node.Host,
		To:      task.User,
		Content: "TaskFinish",
		Body:    task,
	})

	node.Os.LogInfo("stdout", task.Id, "finish", node.GetHostName(), fmt.Sprint(task.CpuRequest), fmt.Sprint(task.MemoryRequest))
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
					From:    node.GetHostName(),
					To:      node.GetHostName(),
					Content: messageType,
					Body:    *t,
				}
				err := node.Os.Send(newMessage)
				if err != nil {
					panic(err)
				}

			}
		}
	}
}
