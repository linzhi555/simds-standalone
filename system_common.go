package main

import (
	"fmt"
	"math/rand"
)

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
		var receivedNum = 0
		for !in.Empty() {
			newM, err := in.Dequeue()
			// message body can not be pointer
			AssertTypeIsNotPointer(newM.Body)
			if IsSameHost(newM.To, newM.From) {
				newM.LeftTime = 0
			} else {
				newM.LeftTime = n.NetLatency
			}

			if err != nil {
				panic(err)
			}

			LogInfo(ecs, entity, ": new message waitting to be send", newM)
			n.Waittings[fmt.Sprint(GetEntityTime(ecs, entity))+"_"+fmt.Sprint(receivedNum)+"_"+newM.From+"_"+newM.To] = &newM
			receivedNum += 1

		}

	}

	for name, v := range n.Waittings {
		if v.LeftTime == 0 {
			LogInfo(ecs, entity, ": new message sended", v)
			out, ok := n.Outs[v.To]
			if !ok {
				panic(v.To + ":net can not reach")
			}
			out.InQueue(*v)
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
			newTask := newMessage.Body.(TaskInfo)
			newTask.StartTime = hostTime
			rm.Tasks[newTask.Id] = &newTask
			newTask.Status = "start"
			LogInfo(ecs, entity, rm.Net.Addr, "Start task:", newTask)
			TaskEventLog(hostTime, &newTask, entity)
		}
	}

	for id, t := range rm.Tasks {
		if t.Status == "start" && t.StartTime+t.LifeTime < GetEntityTime(ecs, entity) {
			t.Status = "finish"
			LogInfo(ecs, entity, rm.Net.Addr, "Task Finished", t)
			TaskEventLog(hostTime, t, entity)
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

const STaskGenUpdate SystemName = "TaskGenUpdateSystem"

func init() { addCommonSystem(STaskGenUpdate, TaskGenUpdateSystem) }
func TaskGenUpdateSystem(ecs *ECS) {
	ecs.ApplyToAllComponent(CTaskGen, TaskGenTicks)
}

func TaskGenTicks(ecs *ECS, entity EntityName, c Component) {
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

	taskNumPerSecond := 0.6 * float32(*NodeNum)
	period := int32(float32(1*Second) / taskNumPerSecond)
	if t%(period) == 2 && t < 10*Second {
		dstAddr := taskgen.Receivers[taskgen.CurTaskId%(len(taskgen.Receivers))]

		newtask := TaskInfo{
			Id:            fmt.Sprintf("task%d", taskgen.CurTaskId),
			CpuRequest:    1 + int32(rand.Intn(4)),
			MemoryRequest: 1 + int32(rand.Intn(4)),
			LifeTime:      (1000 + int32(rand.Intn(5000))) * MiliSecond,
			Status:        "submit",
		}

		newMessage := Message{
			From:    taskgen.Net.Addr,
			To:      dstAddr,
			Content: "TaskDispense",
			Body:    newtask,
		}
		taskgen.Net.Out.InQueue(newMessage)
		TaskEventLog(t, &newtask, entity)
		LogInfo(ecs, entity, fmt.Sprintf(": send task to %s %v", dstAddr, newMessage.Body))
		taskgen.CurTaskId += 1
	}
}
