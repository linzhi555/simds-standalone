package main

import (
	"fmt"
	"math/rand"
)

func AddAllsystemToEcs(e *ECS) {
	e.AddSystem("SystemTimeUpdate", SystemTimeUpdateSystem)
	e.AddSystem("NetworkUpdate", NetworkUpdateSystem)
	//e.AddSystem("NetcardUpdateSystem", NetcardUpdateSystem)
	e.AddSystem("TaskGenUpdateSystem", TaskGenUpdateSystem)
	e.AddSystem("SchedulerUpdateSystem", SchedulerUpdateSystem)
	e.AddSystem("ResourceManagerUpdateSystem", ResourceManagerUpdateSystem)
}

func SystemTimeUpdateSystem(e *ECS) {
	for _, Components := range e.Entities {
		for componentName, Component := range Components {
			if componentName == "SystemTime" {
				st := Component.(*SystemTime)
				st.MicroSecond += 1
			}
		}
	}
}

func NetworkUpdateSystem(ecs *ECS) {
	ecs.ComponentTick("Network", NetworkUpdate)
}

func LogInfo(ecs *ECS, entity string, ins ...interface{}) {
	fmt.Print(GetEntityTime(ecs, entity), " ", "Info", " ", entity, " ")
	for _, item := range ins {
		fmt.Print(item, " ")
	}
	fmt.Println()
}

func NetworkUpdate(ecs *ECS, entity string, c Component) {

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

func GetEntityTime(ecs *ECS, entity string) int32 {
	timeComponent, ok := ecs.GetComponetOfEntity(entity, "SystemTime")
	if ok != true {
		panic("the " + fmt.Sprint(entity) + " should have time component")
	}
	return timeComponent.(*SystemTime).MicroSecond
}

func TaskGenUpdateSystem(ecs *ECS) {
	ecs.ComponentTick("TaskGen", TaskGenTicks)
}
func TaskGenTicks(ecs *ECS, entity string, c Component) {
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
				LifeTime:      1000,
			},
		})
	}
}

func SchedulerUpdateSystem(ecs *ECS) {
	ecs.ComponentTick("Scheduler", SchedulerTicks)
}
func SchedulerTicks(ecs *ECS, entity string, c Component) {
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
		//scheduler.Net.Out.InQueue(newMessage)
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

func ResourceManagerUpdateSystem(ecs *ECS) {
	ecs.ComponentTick("ResourceManager", ResourceManagerTicks)
}
func ResourceManagerTicks(ecs *ECS, entity string, c Component) {
	rm := c.(*ResourceManager)
	hostTime := GetEntityTime(ecs, entity)
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

func UpdateNodeInfo(ecs *ECS, entity string, cpu, memory int32) {
	c, ok := ecs.GetComponetOfEntity(entity, "NodeInfo")
	if ok != true {
		panic("the " + fmt.Sprint(entity) + " should have Node component")
	}

	nodeinfo := c.(*NodeInfo)

	if nodeinfo.CpuAllocted != cpu || nodeinfo.MemoryAllocted != memory {
		nodeinfo.CpuAllocted = cpu
		nodeinfo.MemoryAllocted = memory
		LogInfo(ecs, entity, ":node resource status", cpu, memory)
	}
}
