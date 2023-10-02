package main

import (
	"fmt"
	//"math/rand"
)

var dcssSystem []System

func addDcssSystem(n SystemName, f func(*ECS)) {
	dcssSystem = append(dcssSystem, System{n, f})
}

func RegisteDcssSystemToEcs(e *ECS) {
	for _, s := range commonSystem {
		e.AddSystem(s.Name, s.Function)
	}
	for _, s := range dcssSystem {
		e.AddSystem(s.Name, s.Function)
	}
}

const SDcssSchedulerUpdate = "DcssSchedulerUpdateSystem"

func init() { addDcssSystem(SDcssSchedulerUpdate, DcssSchedulerUpdateSystem) }
func DcssSchedulerUpdateSystem(ecs *ECS) {
	ecs.ApplyToAllComponent(CScheduler, DcssSchedulerTicks)
}
func DcssSchedulerTicks(ecs *ECS, entity EntityName, comp Component) Component {
	scheduler := comp.(Scheduler)
	timeNow := GetEntityTime(ecs, entity)
	//rm := ecs.GetComponet(entity, CResouceManger).(ResourceManager)

	if timeNow == 1 {
		keys := make([]string, 0, len(scheduler.Workers))
		for k := range scheduler.Workers {
			keys = append(keys, k)
		}
		LogInfo(ecs, entity, scheduler.Net.Addr, fmt.Sprintf("Now,I have %d neibor, they are %s", len(scheduler.Workers), keys))
	}

	for !scheduler.Net.In.Empty() {
		newMessage, err := scheduler.Net.In.Dequeue()
		if err != nil {
			panic(err)
		}
		LogInfo(ecs, entity, scheduler.Net.Addr, "received", newMessage.Content, newMessage.Body)

		//task := newMessage.Body.(TaskInfo)

		//switch newMessage.Content {
		//case "TaskDispense":
		//	task.InQueneTime = timeNow
		//	task.Status = "Scheduling"
		//	scheduler.Tasks[task.Id] = &task
		//	// judge if we can run the task locally
		//	if rm.Node.CanAllocate(task.CpuRequest, task.MemoryRequest) {
		//		dstWorker := scheduler.Net.Addr
		//		newMessage := Message{
		//			From:    dstWorker,
		//			To:      rm.Net.Addr,
		//			Content: "TaskAllocate",
		//			Body:    task,
		//		}
		//		task.Status = "Allocated"
		//		scheduler.Net.Out.InQueue(newMessage)
		//	} else {
		//		task.Status = "DiviDeStage1"
		//		task.ScheduleFailCount = 0 // this is for count how many neibor reject this task
		//		keys := make([]string, 0, len(scheduler.Workers))
		//		for k := range scheduler.Workers {
		//			keys = append(keys, k)
		//		}

		//		for _, neibor := range keys {
		//			newMessage := Message{
		//				From:    scheduler.Net.Addr,
		//				To:      neibor,
		//				Content: "TaskDivide",
		//				Body:    task,
		//			}
		//			scheduler.Net.Out.InQueue(newMessage)
		//		}
		//		task.Status = "DiviDeStage2"
		//	}

		//case "TaskDivide":
		//	messageReply := newMessage
		//	messageReply.To = newMessage.From
		//	messageReply.From = newMessage.To
		//	if rm.Node.CanAllocate(task.CpuRequest, task.MemoryRequest) {
		//		messageReply.Content = "TaskDivideConfirm"
		//		scheduler.Tasks[task.Id] = &task
		//		task.Status = "NeedAllocate"
		//	} else {
		//		messageReply.Content = "TaskDivideReject"
		//	}
		//	scheduler.Net.Out.InQueue(messageReply)

		//case "TaskDivideConfirm":
		//	if scheduler.Tasks[task.Id].Status == "DiviDeStage2" {
		//		scheduler.Tasks[task.Id].Status = "DiviDeStage3"
		//		messageReply := newMessage
		//		messageReply.To = newMessage.From
		//		messageReply.From = newMessage.To
		//		messageReply.Body = *scheduler.Tasks[task.Id]
		//		messageReply.Content = "TaskDivideAllocate"
		//		scheduler.Net.Out.InQueue(messageReply)
		//	}

		//case "TaskDivideAllocate":
		//	if scheduler.Tasks[task.Id].Status == "NeedAllocate" {
		//		scheduler.Tasks[task.Id].Status = "Allocate"
		//		messageReply := newMessage
		//		messageReply.To = rm.Net.Addr
		//		messageReply.From = newMessage.To
		//		messageReply.Content = "TaskAllocate"
		//		scheduler.Net.Out.InQueue(messageReply)
		//	}

		//case "TaskDivideReject":
		//	scheduler.Tasks[task.Id].ScheduleFailCount += 1
		//	neiborNum := len(scheduler.Workers)
		//	// if all neibors reject this task, so we i have to dispense the task to a random neibors,
		//	// the distination neibors  may have a valid neibor to execute this task
		//	if scheduler.Tasks[task.Id].ScheduleFailCount == int32(neiborNum) {
		//		var taskCopy TaskInfo = *scheduler.Tasks[task.Id]
		//		neibors := scheduler.GetAllWokersName()
		//		dstNeibor := neibors[rand.Intn(len(neibors))]

		//		newMessage := Message{
		//			From:    scheduler.Net.Addr,
		//			To:      dstNeibor,
		//			Content: "TaskDispense",
		//			Body:    taskCopy,
		//		}
		//		scheduler.Net.Out.InQueue(newMessage)
		//		LogInfo(ecs, entity, scheduler.Net.Addr, "TaskDivide finally fail, start a new TaskDispense", newMessage.Body)
		//	}
		//}
	}
	return scheduler

}
