package main

import (
	"fmt"
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
func DcssSchedulerTicks(ecs *ECS, entity EntityName, c Component) {
	scheduler := c.(*Scheduler)
	timeNow := GetEntityTime(ecs, entity)
	rm := ecs.GetComponet(entity, CResouceManger).(*ResourceManager)
	nodeinf := ecs.GetComponet(entity, CNodeInfo).(*NodeInfo)

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

		task := newMessage.Body.(*TaskInfo).DeepCopy()
		if newMessage.Content == "TaskDispense" {
			task.InQueneTime = timeNow
			task.Status = "Scheduling"
			scheduler.Tasks[task.Id] = task
			// judge if we can run the task locally
			if nodeinf.CanAllocate(task.CpuRequest, task.MemoryRequest) {
				dstWorker := scheduler.Net.Addr
				newMessage := &Message{
					From:    dstWorker,
					To:      rm.Net.Addr,
					Content: "TaskAllocate",
					Body:    task,
				}
				task.Status = "Allocated"
				scheduler.Net.Out.InQueue(newMessage)
			} else {
				task.Status = "DiviDeStage1"

				keys := make([]string, 0, len(scheduler.Workers))
				for k := range scheduler.Workers {
					keys = append(keys, k)
				}

				for _, neibor := range keys {
					newMessage := &Message{
						From:    scheduler.Net.Addr,
						To:      neibor,
						Content: "TaskDivide",
						Body:    task.DeepCopy(),
					}
					scheduler.Net.Out.InQueue(newMessage)
				}
				task.Status = "DiviDeStage2"
			}
		}

		if newMessage.Content == "TaskDivide" {
			messageReply := *newMessage
			messageReply.To = newMessage.From
			messageReply.From = newMessage.To
			if nodeinf.CanAllocate(task.CpuRequest, task.MemoryRequest) {
				messageReply.Content = "TaskDivideConfirm"
				scheduler.Tasks[task.Id] = task
				task.Status = "NeedAllocate"
			} else {
				messageReply.Content = "TaskDivideReject"
			}
			scheduler.Net.Out.InQueue(&messageReply)
		}

		if newMessage.Content == "TaskDivideConfirm" {
			if scheduler.Tasks[task.Id].Status == "DiviDeStage2" {
				scheduler.Tasks[task.Id].Status = "DiviDeStage3"
				messageReply := *newMessage
				messageReply.To = newMessage.From
				messageReply.From = newMessage.To
				messageReply.Body = scheduler.Tasks[task.Id]
				messageReply.Content = "TaskDivideAllocate"
				scheduler.Net.Out.InQueue(&messageReply)
			}
		}

		if newMessage.Content == "TaskDivideAllocate" {
			if scheduler.Tasks[task.Id].Status == "NeedAllocate" {
				scheduler.Tasks[task.Id].Status = "Allocate"
				messageReply := *newMessage
				messageReply.To = rm.Net.Addr
				messageReply.From = newMessage.To
				messageReply.Content = "TaskAllocate"
				scheduler.Net.Out.InQueue(&messageReply)
			}
		}

		if newMessage.Content == "TaskDivideReject" {
			// pass
		}
	}

}
