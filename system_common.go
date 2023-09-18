package main

import "fmt"

var commonSystem []System

func addCommonSystem(n SystemName, f func(*ECS)) {
	commonSystem = append(commonSystem, System{n, f})
}

const SSystemTimeUpdate = "SystemTimeUpdate"

func init() { addCommonSystem(SSystemTimeUpdate, SystemTimeUpdate) }
func SystemTimeUpdate(e *ECS) {
	for _, Components := range e.Entities {
		for componentName, Component := range Components {
			if componentName == "SystemTime" {
				st := Component.(*SystemTime)
				st.Time += 1
			}
		}
	}
}

const SNetworkUpdate = "NetworkUpdateSystem"

func init() { addCommonSystem(SNetworkUpdate, NetworkUpdateSystem) }
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
			out, ok := n.Outs[v.To]
			if !ok {
				panic(v.To + ":net can not reach")
			}
			out.InQueue(v)
			delete(n.Waittings, name)
		} else {
			v.LeftTime -= 1
		}
	}

}

func GetEntityTime(ecs *ECS, entity EntityName) int32 {
	timeComponent := ecs.GetComponet(entity, CSystemTime)
	return timeComponent.(*SystemTime).Time
}

const SResourceManagerUpdate = "ResourceManagerUpdateSystem"

func init() { addCommonSystem(SResourceManagerUpdate, ResourceManagerUpdateSystem) }
func ResourceManagerUpdateSystem(ecs *ECS) {
	ecs.ApplyToAllComponent(CResouceManger, ResourceManagerTicks)
}
func ResourceManagerTicks(ecs *ECS, entity EntityName, c Component) {
	rm := c.(*ResourceManager)
	hostTime := GetEntityTime(ecs, entity)

	if !rm.Net.In.Empty() {
		newMessage, err := rm.Net.In.Dequeue()
		if err != nil {
			panic(err)
		}
		LogInfo(ecs, entity, rm.Net.Addr, "received message:", newMessage)

		if newMessage.Content == "TaskAllocate" {
			newTask := newMessage.Body.(*TaskInfo)
			newTask.StartTime = hostTime
			rm.Tasks[newTask.Id] = newTask
			newTask.Status = "Run"
			LogInfo(ecs, entity, rm.Net.Addr, "Start task:", newTask)
		}

	}

	for id, t := range rm.Tasks {
		if t.Status == "Run" && t.StartTime+t.LifeTime < GetEntityTime(ecs, entity) {
			LogInfo(ecs, entity, rm.Net.Addr, "Task Finished", t)
			delete(rm.Tasks, id)
		}
	}

	var allcpu int32 = 0
	var allmemory int32 = 0

	for _, t := range rm.Tasks {
		allcpu += t.CpuRequest
		allmemory += t.MemoryRequest
	}

	if hostTime%(5*MiliSecond) == 1 {
		UpdateNodeInfo(ecs, entity, allcpu, allmemory)
	}
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
