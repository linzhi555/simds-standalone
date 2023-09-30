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
			if rm.TaskFinishReceiver != "" {
				informReceiverTaskStatus(rm, t, "TaskFinish")
			}

			delete(rm.Tasks, id)
		}
	}

	updateNodeInfo(ecs, entity, rm)
}

func informReceiverTaskStatus(rm *ResourceManager, t *TaskInfo, content string) {
	newMessage := Message{
		From:    rm.Net.Addr,
		To:      rm.TaskFinishReceiver,
		Content: content,
		Body:    *t,
	}
	rm.Net.Out.InQueue(newMessage)
}

func updateNodeInfo(ecs *ECS, entity EntityName, rm *ResourceManager) {
	var cpu int32 = 0
	var memory int32 = 0

	for _, t := range rm.Tasks {
		cpu += t.CpuRequest
		memory += t.MemoryRequest
	}

	if rm.Node.CpuAllocted != cpu || rm.Node.MemoryAllocted != memory {
		rm.Node.CpuAllocted = cpu
		rm.Node.MemoryAllocted = memory
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

	taskNumPerSecond := Config.TaskNumFactor * float32(Config.NodeNum)
	if t < 10*Second {
		for taskgen.CurTaskId < int(taskNumPerSecond*float32(t)/float32(Second)) {
			dstAddr := taskgen.Receivers[taskgen.CurTaskId%(len(taskgen.Receivers))]

			newtask := TaskInfo{
				Id:            fmt.Sprintf("task%d", taskgen.CurTaskId),
				CpuRequest:    1 + int32(rand.Intn(int(2*Config.TaskCpu-2))),
				MemoryRequest: 1 + int32(rand.Intn(int(2*Config.TaskMemory-2))),
				LifeTime:      (100 + int32(rand.Intn(int(Config.TaskLifeTime)*2-200))) * MiliSecond,
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
}
