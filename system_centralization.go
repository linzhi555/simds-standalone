package main

import (
	"math/rand"
)

var centralizedSystem []System

func addCentralizedSystem(n SystemName, f func(*ECS)) {
	centralizedSystem = append(centralizedSystem, System{n, f})
}
func RegisteCentralizedsystemToEcs(e *ECS) {
	for _, s := range commonSystem {
		e.AddSystem(s.Name, s.Function)
	}
	for _, s := range centralizedSystem {
		e.AddSystem(s.Name, s.Function)
	}
}

const SSchedulerUpdate = "SchedulerUpdateSystem"

func init() { addCentralizedSystem(SSchedulerUpdate, SchedulerUpdateSystem) }
func SchedulerUpdateSystem(ecs *ECS) {
	ecs.ApplyToAllComponent("Scheduler", SchedulerTicks)
}
func SchedulerTicks(ecs *ECS, entity EntityName, c Component) {
	scheduler := c.(*Scheduler)
	timeNow := GetEntityTime(ecs, entity)

	for !scheduler.Net.In.Empty() {
		newMessage, err := scheduler.Net.In.Dequeue()
		if err != nil {
			panic(err)
		}

		if newMessage.Content == "TaskDispense" {
			task := newMessage.Body.(TaskInfo)
			task.InQueneTime = timeNow
			task.Status = "Scheduling"
			scheduler.Tasks[task.Id] = &task
			LogInfo(ecs, entity, scheduler.Net.Addr, "received task submit", task)
		}

		if newMessage.Content == "WorkerUpdate" {
			nodeinfo := newMessage.Body.(NodeInfo)
			scheduler.Workers[newMessage.From] = &(nodeinfo)
			LogInfo(ecs, entity, scheduler.Net.Addr, "received WorkerUpdate", newMessage.From, nodeinfo)
		}

	}

	for _, task := range scheduler.Tasks {
		if task.Status == "Scheduling" {
			if timeNow-task.InQueneTime > 10*MiliSecond {
				task.Status = "Scheduled"
			}

		}
		if task.Status == "Scheduled" {
			dstWorker := schdulingAlgorithm(scheduler)
			newMessage := Message{
				From:    scheduler.Net.Addr,
				To:      dstWorker,
				Content: "TaskAllocate",
				Body:    *task,
			}
			scheduler.Net.Out.InQueue(newMessage)
			task.Status = "Allocated"
			LogInfo(ecs, entity, scheduler.Net.Addr, "sendtask to", dstWorker, task)
		}
	}

}

func schdulingAlgorithm(scheduler *Scheduler) string {
	keys := make([]string, 0, len(scheduler.Workers))
	for k := range scheduler.Workers {
		keys = append(keys, k)
	}

	addr := keys[rand.Intn(len(keys))]
	return addr
}

const SWorkerStatusUpdate = "WorkerStatusUpdateSystem"

func init() { addCentralizedSystem(SWorkerStatusUpdate, WorkerStatusUpdateSystem) }
func WorkerStatusUpdateSystem(ecs *ECS) {
	ecs.ApplyToAllComponent(CResouceManger, WorkerStatusUpdateTicks)
}
func WorkerStatusUpdateTicks(ecs *ECS, entity EntityName, c Component) {
	rm := c.(*ResourceManager)
	hostTime := GetEntityTime(ecs, entity)
	tmp := ecs.GetComponet(entity, CNodeInfo)
	nodeinfo := tmp.(*NodeInfo)

	if hostTime%(1000*MiliSecond) == 1 {
		nodeinfoCopy := *nodeinfo
		rm.Net.Out.InQueue(Message{
			From:    rm.Net.Addr,
			To:      "master1:Scheduler",
			Content: "WorkerUpdate",
			Body:    nodeinfoCopy,
		})
		LogInfo(ecs, entity, rm.Net.Addr, "WorkerUpdate", nodeinfoCopy)
	}

}
