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
	taskgenComponet := c.(*TaskGen)

	period := 10 * MiliSecond
	if t%(period) == 1 && t < 10*Second {
		dstAddr := fmt.Sprintf("node%d:Scheduler", taskgenComponet.CurTaskId%10)
		newMessage := &Message{
			From:    taskgenComponet.Net.Addr,
			To:      dstAddr,
			Content: "TaskDispense",
			Body: &TaskInfo{
				Id:            fmt.Sprintf("task%d", taskgenComponet.CurTaskId),
				CpuRequest:    1,
				MemoryRequest: 1,
				LifeTime:      500 * MiliSecond,
			},
		}
		taskgenComponet.Net.Out.InQueue(newMessage)
		LogInfo(ecs, entity, fmt.Sprintf(": send task to %s %v", dstAddr, newMessage.Body))
		taskgenComponet.CurTaskId += 1
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

	if timeNow == 1 {
		keys := make([]string, 0, len(scheduler.Workers))
		for k := range scheduler.Workers {
			keys = append(keys, k)
		}
		LogInfo(ecs, entity, scheduler.Net.Addr, fmt.Sprintf("Now,I have %d neibor, they are %s", len(scheduler.Workers), keys))
	}

	rm := ecs.GetComponet(entity, CResouceManger).(*ResourceManager)
	nodeinf := ecs.GetComponet(entity, CNodeInfo).(*NodeInfo)
	if !scheduler.Net.In.Empty() {
		newMessage, err := scheduler.Net.In.Dequeue()
		if err != nil {
			panic(err)
		}

		if newMessage.Content == "TaskDispense" {
			task := newMessage.Body.(*TaskInfo)
			task.InQueneTime = timeNow
			task.Status = "Scheduling"
			scheduler.Tasks[task.Id] = task
			LogInfo(ecs, entity, scheduler.Net.Addr, "received TaskDispense", task)
		}

		if newMessage.Content == "TaskDivide" {
			task := newMessage.Body.(*TaskInfo)
			task.InQueneTime = timeNow
			task.Status = "Scheduling"
			scheduler.Tasks[task.Id] = task
			LogInfo(ecs, entity, scheduler.Net.Addr, "received TaskDivide", task)
		}

		if newMessage.Content == "WorkerUpdate" {
			nodeinfo := newMessage.Body.(*NodeInfo)
			scheduler.Workers[newMessage.From] = &(*nodeinfo)
			LogInfo(ecs, entity, scheduler.Net.Addr, "received WorkerUpdate", newMessage.From, *nodeinfo)
		}

	}

	for _, task := range scheduler.Tasks {
		if task.Status == "Scheduling" {
			if timeNow-task.InQueneTime > 10*MiliSecond {
				task.Status = "Scheduled"
			}

		}
		if task.Status == "Scheduled" {
			var dstWorker string
			var actionType string
			if nodeinf.CanAllocate(task.CpuRequest, task.MemoryRequest) {
				dstWorker = rm.Net.Addr
				actionType = "TaskAllocate"
			} else {
				keys := make([]string, 0, len(scheduler.Workers))
				for k := range scheduler.Workers {
					keys = append(keys, k)
				}
				dstWorker = keys[rand.Intn(len(keys))]
				actionType = "TaskDivide"
			}

			newMessage := &Message{
				From:    scheduler.Net.Addr,
				To:      dstWorker,
				Content: actionType,
				Body:    task,
			}

			scheduler.Net.Out.InQueue(newMessage)
			if actionType == "TaskAllocate" {
				task.Status = "Allocated"
				LogInfo(ecs, entity, scheduler.Net.Addr, "Run task locally", dstWorker, task)
			} else {
				task.Status = "Divided"
				LogInfo(ecs, entity, scheduler.Net.Addr, "Divide the task", dstWorker, task)

			}

		}
	}
}
