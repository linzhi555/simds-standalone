package main

import "fmt"

func AddAllsystemToEcs(e *ECS) {
	e.AddSystem("SystemTimeUpdate", SystemTimeUpdateSystem)
	e.AddSystem("NetworkUpdate", NetworkUpdateSystem)
	e.AddSystem("NetcardUpdateSystem", NetcardUpdateSystem)
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
	for !n.Ins.Empty() {
		newM, err := n.Ins.Dequeue()
		if err != nil {
			panic(err)
		}
		n.Waittings[newM.From] = newM

	}

	for from, v := range n.Waittings {
		if v == nil {
			return
		}
		if v.LeftTime == 0 {
			n.Outs[v.To].InQueue(v)
			delete(n.Waittings, from)
		} else {
			v.LeftTime -= 1
		}
	}

}

func NetcardUpdateSystem(ecs *ECS) {
	ecs.ComponentTick("NetCard", NetCardTicks)
}

func NetCardTicks(ecs *ECS, entity string, c Component) {
	nc := c.(*NetCard)
	if nc.In.Empty() != true {
		newMessage, _ := nc.In.Dequeue()
		fmt.Println("receive new message", newMessage)
	}
}
