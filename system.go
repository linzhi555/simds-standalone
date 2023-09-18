package main

import (
	"fmt"
	"math/rand"
)

var allSystem []System

func addSystem(n SystemName, f func(*ECS)) {
	allSystem = append(allSystem, System{n, f})
}
func RegisteAllsystemToEcs(e *ECS) {
	for _, s := range allSystem {
		e.AddSystem(s.Name, s.Function)
	}
}

const SSystemTimeUpdate = "SystemTimeUpdate"

func init() { addSystem(SSystemTimeUpdate, SystemTimeUpdate) }
func SystemTimeUpdate(e *ECS) {
	for _, Components := range e.Entities {
		for componentName, Component := range Components {
			if componentName == "SystemTime" {
				st := Component.(*SystemTime)
				st.MicroSecond += 1
			}
		}
	}
}

const SNetworkUpdate = "NetworkUpdateSystem"

func init() { addSystem(SNetworkUpdate, NetworkUpdateSystem) }
func NetworkUpdateSystem(ecs *ECS) {
	ecs.ApplyToAllComponent("Network", NetworkUpdate)
}
func NetworkUpdate(ecs *ECS, entity EntityName, c Component) {

	n := c.(*Network)

	for _, in := range n.Ins {
		for !in.Empty() {
			newM, err := in.Dequeue()
			newM.LeftTime = n.NetLatency
			if err != nil {
				panic(err)
			}

			LogInfo(ecs, entity, ": new message waitting to be send", newM)
			n.Waittings[fmt.Sprint(GetEntityTime(ecs, entity))+"_"+newM.From] = newM

		}

	}

	for name, v := range n.Waittings {
		if v.LeftTime == 0 {
			LogInfo(ecs, entity, ": new message sended", v)
			n.Outs[v.To].InQueue(v)
			delete(n.Waittings, name)
		} else {
			v.LeftTime -= 1
		}
	}

}

func GetEntityTime(ecs *ECS, entity EntityName) int32 {
	timeComponent := ecs.GetComponet(entity, CSystemTime)
	return timeComponent.(*SystemTime).MicroSecond
}

const STaskGenUpdate = "TaskGenUpdateSystem"

func init() { addSystem(STaskGenUpdate, TaskGenUpdateSystem) }
func TaskGenUpdateSystem(ecs *ECS) {
	ecs.ApplyToAllComponent(CTaskGen, TaskGenTicks)
}
func TaskGenTicks(ecs *ECS, entity EntityName, c Component) {
	t := GetEntityTime(ecs, entity)
	taskgenComponet := c.(*TaskGen)
	if t%100 == 1 && t < 10000 {
		LogInfo(ecs, entity, " : send task to master1:Scheduler ")
		taskgenComponet.Net.Out.InQueue(&Message{
			From:    taskgenComponet.Net.Addr,
			To:      "master1:Scheduler",
			Content: "TaskSubmit",
			Body: &TaskInfo{
				Id:            fmt.Sprintf("task%d", t/100),
				CpuRequest:    1,
				MemoryRequest: 1,
				LifeTime:      2000,
			},
		})
	}
}

const SSchedulerUpdate = "SchedulerUpdateSystem"

func init() { addSystem(SSchedulerUpdate, SchedulerUpdateSystem) }
func SchedulerUpdateSystem(ecs *ECS) {
	ecs.ApplyToAllComponent("Scheduler", SchedulerTicks)
}
func SchedulerTicks(ecs *ECS, entity EntityName, c Component) {
	scheduler := c.(*Scheduler)
	timeNow := GetEntityTime(ecs, entity)

	if !scheduler.Net.In.Empty() {
		newMessage, err := scheduler.Net.In.Dequeue()
		if err != nil {
			panic(err)
		}

		if newMessage.Content == "TaskSubmit" {
			task := newMessage.Body.(*TaskInfo)
			task.InQueneTime = timeNow
			task.Status = "Scheduling"
			scheduler.Tasks[task.Id] = task
			LogInfo(ecs, entity, scheduler.Net.Addr, "received task submit", task)
		}

		if newMessage.Content == "WorkerUpdate" {
			nodeinfo := newMessage.Body.(*NodeInfo)
			scheduler.Workers[newMessage.From] = &(*nodeinfo)
			LogInfo(ecs, entity, scheduler.Net.Addr, "received WorkerUpdate", newMessage.From, *nodeinfo)
		}

	}

	for _, task := range scheduler.Tasks {
		if task.Status == "Scheduling" {
			if timeNow-task.InQueneTime > 500 {
				task.Status = "Scheduled"
			}

		}
		if task.Status == "Scheduled" {
			dstWorker := schdulingAlgorithm(scheduler)
			newMessage := &Message{
				From:    scheduler.Net.Addr,
				To:      dstWorker,
				Content: "TaskDispense",
				Body:    task,
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

const SResourceManagerUpdate = "ResourceManagerUpdateSystem"

func init() { addSystem(SResourceManagerUpdate, ResourceManagerUpdateSystem) }
func ResourceManagerUpdateSystem(ecs *ECS) {
	ecs.ApplyToAllComponent(CResouceManger, ResourceManagerTicks)
}
func ResourceManagerTicks(ecs *ECS, entity EntityName, c Component) {
	rm := c.(*ResourceManager)
	hostTime := GetEntityTime(ecs, entity)
	tmp := ecs.GetComponet(entity, CNodeInfo)
	nodeinfo := tmp.(*NodeInfo)

	if !rm.Net.In.Empty() {
		newMessage, err := rm.Net.In.Dequeue()
		if err != nil {
			panic(err)
		}

		LogInfo(ecs, entity, rm.Net.Addr, "received message:", newMessage)
		newTask := newMessage.Body.(*TaskInfo)
		newTask.StartTime = hostTime
		rm.Tasks[newTask.Id] = newTask
	}

	if hostTime%1000 == 99 {
		nodeinfoCopy := *nodeinfo
		rm.Net.Out.InQueue(&Message{
			From:    rm.Net.Addr,
			To:      "master1:Scheduler",
			Content: "WorkerUpdate",
			Body:    &nodeinfoCopy,
		})
		LogInfo(ecs, entity, rm.Net.Addr, "WorkerUpdate", nodeinfoCopy)
	}

	for id, t := range rm.Tasks {
		if t.StartTime+t.LifeTime < GetEntityTime(ecs, entity) {
			delete(rm.Tasks, id)
		}
	}

	var allcpu int32 = 0
	var allmemory int32 = 0

	for _, t := range rm.Tasks {
		allcpu += t.CpuRequest
		allmemory += t.MemoryRequest
	}

	UpdateNodeInfo(ecs, entity, allcpu, allmemory)
}

func UpdateNodeInfo(ecs *ECS, entity EntityName, cpu, memory int32) {
	c := ecs.GetComponet(entity, CNodeInfo)
	nodeinfo := c.(*NodeInfo)

	if nodeinfo.CpuAllocted != cpu || nodeinfo.MemoryAllocted != memory {
		nodeinfo.CpuAllocted = cpu
		nodeinfo.MemoryAllocted = memory
		LogInfo(ecs, entity, ":node resource status", cpu, memory)
	}
}
