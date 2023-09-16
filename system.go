package main

import "fmt"

func AddAllsystemToEcs(e *ECS) {
	e.AddSystem("SystemTimeUpdate", SystemTimeUpdateSystem)
	e.AddSystem("NetworkUpdate", NetworkUpdateSystem)
	//e.AddSystem("NetcardUpdateSystem", NetcardUpdateSystem)
	e.AddSystem("TaskGenUpdateSystem", TaskGenUpdateSystem)
	e.AddSystem("SchedulerUpdateSystem", SchedulerUpdateSystem)
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

func NetworkUpdate(ecs *ECS, entity string, c Component) {

	n := c.(*Network)

	for _, in := range n.Ins {
		for !in.Empty() {
			newM, err := in.Dequeue()
			newM.LeftTime = 5
			if err != nil {
				panic(err)
			}

			fmt.Println(GetEntityTime(ecs, entity), "new message waitting to be send", newM)
			n.Waittings[fmt.Sprint(GetEntityTime(ecs, entity))+"_"+newM.From] = newM

		}

	}

	for name, v := range n.Waittings {
		if v.LeftTime == 0 {
			fmt.Println(GetEntityTime(ecs, entity), ": newwork message sended", v)
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

func NetcardUpdateSystem(ecs *ECS) {
	ecs.ComponentTick("NetCard", NetCardTicks)
}

func NetCardTicks(ecs *ECS, entity string, c Component) {
	nc := c.(*NetCard)
	if nc.In.Empty() != true {
		newMessage, _ := nc.In.Dequeue()
		fmt.Println(GetEntityTime(ecs, entity), "receive new message", newMessage)
	}
}
func TaskGenUpdateSystem(ecs *ECS) {
	ecs.ComponentTick("TaskGen", TaskGenTicks)
}
func TaskGenTicks(ecs *ECS, entity string, c Component) {
	t := GetEntityTime(ecs, entity)
	taskgenComponet := c.(*TaskGen)
	if t%2 == 1 && t < 10 {
		fmt.Println(t, "send task to master1:Scheduler ")
		taskgenComponet.Net.Out.InQueue(&Message{
			From:    taskgenComponet.Net.Addr,
			To:      "master1:Scheduler",
			Content: "TaskSubmit",
			Body: &TaskInfo{
				CpuRequest:    1,
				MemoryRequest: 1,
			},
		})
	}
}

func SchedulerUpdateSystem(ecs *ECS) {
	ecs.ComponentTick("Scheduler", SchedulerTicks)
}
func SchedulerTicks(ecs *ECS, entity string, c Component) {
	scheduler := c.(*Scheduler)
	if !scheduler.Net.In.Empty() {
		newMessage, err := scheduler.Net.In.Dequeue()
		if err != nil {
			panic(err)
		}
		fmt.Println(GetEntityTime(ecs, entity), scheduler.Net.Addr, "received message:", newMessage)

	}

}
