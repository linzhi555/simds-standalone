package main

import (
	"fmt"
	"math/rand"
)

func BuildDCSSCluster() Cluster {
	var cluster = createCluster()
	var nodes []Node
	nodes = append(nodes, Node{
		"user1",
		[]NodeComponent{
			NewTaskGen("user1"),
		},
	})

	for i := 0; i < int(Config.NodeNum); i++ {

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

func DcssTaskgenSetup(c interface{}) {
	taskgen := c.(*TaskGen)
	taskgen.StartTime = taskgen.Os.GetTime()
	for i := 0; i < int(Config.NodeNum); i++ {

		taskgen.Receivers = append(taskgen.Receivers,
			"node"+fmt.Sprint(i)+":"+string(CScheduler),
		)
	}

}
func DcssSchedulerSetup(comp interface{}) {
	scheduler := comp.(*Scheduler)

	var neibors []string
	// add self in the first for convernience, and ignore the first when actually register neibor
	neiborNum := int(Config.DcssNeibor)
	allNodeNum := int(Config.NodeNum)
	neibors = append(neibors, scheduler.Os.Net().GetAddr())
	for len(neibors) != neiborNum+1 {
		newNeibor := fmt.Sprintf("node%d:Scheduler", rand.Intn(allNodeNum))
		alreadyExisted := false
		for _, n := range neibors {
			if n == newNeibor {
				alreadyExisted = true
			}
		}
		if !alreadyExisted {
			neibors = append(neibors, newNeibor)
		}
	}

	for _, n := range neibors[1:] {

		nodeInfo := &NodeInfo{n, Config.NodeCpu, Config.NodeMemory, 0, 0}
		scheduler.Workers[n] = nodeInfo.Clone()
	}
	scheduler.LocalNode = &NodeInfo{scheduler.Os.Net().GetAddr(), Config.NodeCpu, Config.NodeMemory, 0, 0}
	keys := make([]string, 0, len(scheduler.Workers))
	for k := range scheduler.Workers {
		keys = append(keys, k)
	}
	LogInfo(scheduler.Os, fmt.Sprintf("Now,I have %d neibor, they are %s", len(scheduler.Workers), keys))

}

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
}

func dcssTaskDispenseHandle(scheduler *Scheduler, newMessage Message) {
	task := newMessage.Body.(TaskInfo)
	task.Status = "Scheduling"
	if scheduler.LocalNode.CanAllocateTask(&task) {
		task.Status = "allocate"
		scheduler.LocalNode.AddAllocated(task.CpuRequest, task.MemoryRequest)
		dcssChangeTaskStatusLocally(scheduler, &task, "TaskRun")
		LogInfo(scheduler.Os, "run task locally", task)
	} else {

		LogInfo(scheduler.Os, "start divide ", task)
		task.Status = "DiviDeStage1"
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
			scheduler.Os.Net().Send(newMessage)
		}
		task.Status = "DiviDeStage2"
	}
	scheduler.TasksStatus[task.Id] = &task

}
func dcssTaskDivideHandle(scheduler *Scheduler, newMessage Message) {
	task := newMessage.Body.(TaskInfo)
	messageReply := newMessage
	messageReply.To = newMessage.From
	messageReply.From = newMessage.To
	if scheduler.LocalNode.CanAllocateTask(&task) {
		messageReply.Content = "TaskDivideConfirm"
		task.Status = "needStart"
		scheduler.LocalNode.AddAllocated(task.CpuRequest, task.MemoryRequest)
		dcssChangeTaskStatusLocally(scheduler, &task, "TaskPreAllocate")
		scheduler.TasksStatus[task.Id] = &task
	} else {
		messageReply.Content = "TaskDivideReject"
	}
	scheduler.Os.Net().Send(messageReply)
}
func dcssTaskDivideConfirmHandle(scheduler *Scheduler, newMessage Message) {
	task := newMessage.Body.(TaskInfo)
	t := scheduler.TasksStatus[task.Id]

	if t.Status == "DiviDeStage2" {
		t.Status = "DiviDeStage3"
		scheduler.Os.Net().Send(Message{
			From:    newMessage.To,
			To:      newMessage.From,
			Content: "TaskDivideAllocate",
			Body:    *scheduler.TasksStatus[task.Id],
		})
	} else if t.Status == "DiviDeStage3" {
		scheduler.Os.Net().Send(Message{
			From:    newMessage.To,
			To:      newMessage.From,
			Content: "TaskDivideCancel",
			Body:    *scheduler.TasksStatus[task.Id],
		})
	}
}
func dcssTaskDivideAllocateHandle(scheduler *Scheduler, newMessage Message) {
	task := newMessage.Body.(TaskInfo)
	if scheduler.TasksStatus[task.Id].Status == "needStart" {
		dcssChangeTaskStatusLocally(scheduler, &task, "TaskStart")
	}
}

func dcssTaskDivideCancelHandle(scheduler *Scheduler, newMessage Message) {
	task := newMessage.Body.(TaskInfo)
	if scheduler.TasksStatus[task.Id].Status == "needStart" {
		dcssChangeTaskStatusLocally(scheduler, &task, "TaskCancelAllocate")
	}
}

func dcssTaskDivideRejectHandle(scheduler *Scheduler, newMessage Message) {
	task := newMessage.Body.(TaskInfo)
	scheduler.TasksStatus[task.Id].ScheduleFailCount += 1
	neiborNum := Config.DcssNeibor
	// if all neibors reject this task, so we i have to dispense the task to a random neibors,
	// the distination neibors  may have a valid neibor to execute this task
	if scheduler.TasksStatus[task.Id].ScheduleFailCount == int32(neiborNum) {
		var taskCopy TaskInfo = *scheduler.TasksStatus[task.Id]
		neibors := scheduler.GetAllWokersName()
		dstNeibor := neibors[rand.Intn(len(neibors))]

		newMessage := Message{
			From:    scheduler.Os.Net().GetAddr(),
			To:      dstNeibor,
			Content: "TaskDispense",
			Body:    taskCopy,
		}
		scheduler.Os.Net().Send(newMessage)
		LogInfo(scheduler.Os, "TaskDivide finally fail, start a new TaskDispense", newMessage.Body)
	}
}

func dcssChangeTaskStatusLocally(scheduler *Scheduler, task *TaskInfo, contentType string) {
	dstWorker := scheduler.Host + ":" + string(CResouceManger)
	newMessage := Message{
		From:    scheduler.Os.Net().GetAddr(),
		To:      dstWorker,
		Content: contentType,
		Body:    *task,
	}
	scheduler.Os.Net().Send(newMessage)
}

func dcssFinishHandle(scheduler *Scheduler, newMessage Message) {
	task := newMessage.Body.(TaskInfo)
	scheduler.LocalNode.SubAllocated(task.CpuRequest, task.MemoryRequest)
}

func DcssResourceManagerSetup(comp interface{}) {
	rm := comp.(*ResourceManager)
	rm.TaskFinishReceiver = rm.Host + ":" + string(CScheduler)
}
