package main

import (
	"fmt"
	"math/rand"
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

const SDcssTaskGenUpdate = "DcssTaskGenUpdateSystem"

func init() { addDcssSystem(SDcssTaskGenUpdate, DcssTaskGenUpdateSystem) }
func DcssTaskGenUpdateSystem(ecs *ECS) {
	ecs.ApplyToAllComponent(CTaskGen, DcssTaskGenTicks)
}

func DcssTaskGenTicks(ecs *ECS, entity EntityName, c Component) {
	t := GetEntityTime(ecs, entity)
	taskgen := c.(*TaskGen)

	if t == 1 {
		nodes := ecs.GetEntitiesHasComponenet(CScheduler)
		for _, n := range nodes {
			newReceiver := string(n) + ":" + "Scheduler"
			taskgen.Receivers = append(taskgen.Receivers, newReceiver)
			LogInfo(ecs, entity, fmt.Sprintf(": newReceiver %s", newReceiver))
		}
		return
	}

	period := 10 * MiliSecond
	if t%(period) == 2 && t < 10*Second {
		dstAddr := taskgen.Receivers[taskgen.CurTaskId%(len(taskgen.Receivers))]

		newtask := &TaskInfo{
			Id:            fmt.Sprintf("task%d", taskgen.CurTaskId),
			CpuRequest:    1 + int32(rand.Intn(4)),
			MemoryRequest: 1 + int32(rand.Intn(4)),
			LifeTime:      (1000 + int32(rand.Intn(5000))) * MiliSecond,
			Status:        "submit",
		}

		newMessage := &Message{
			From:    taskgen.Net.Addr,
			To:      dstAddr,
			Content: "TaskDispense",
			Body:    newtask,
		}
		taskgen.Net.Out.InQueue(newMessage)
		TaskEventLog(t, newtask, entity)
		LogInfo(ecs, entity, fmt.Sprintf(": send task to %s %v", dstAddr, newMessage.Body))
		taskgen.CurTaskId += 1
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
