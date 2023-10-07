package main

import (
	"fmt"
	"log"
	"simds-standalone/ecs"
	"time"
)

var ZEROTIME time.Time = time.Now()

type Node struct {
	Name       string
	Components []NodeComponent
}

func EcsRunCluster(nodes []Node) {
	simulator := ecs.NewEcs()

	newNet := CreateMockNetWork(Config.NetLatency * MiliSecond)
	newNet.Net()
	simulator.AddEntities("network1", newNet)

	for _, node := range nodes {
		var inited []ecs.Component
		for _, c := range node.Components {
			c.InitTimeGetter(func() time.Time { return ZEROTIME.Add(time.Duration(simulator.UpdateCount) * time.Millisecond) })

			card := CreateMockNetCard(node.Name + ":" + string(c.Component()))
			card.JoinNetWork(&newNet)
			c.InitNet(card)

			inited = append(inited, c)
		}
		simulator.AddEntities(ecs.EntityName(node.Name), inited...)
	}

	simulator.AddSystem("taskgenUpdate",func(e *ecs.ECS){
		f := func(ecs *ecs.ECS, e ecs.EntityName, comp ecs.Component) ecs.Component {
			t := comp.(TaskGen)
			t.Net().Send(Message{From:"",To: "worker1:ResourceManager",Content: "asdf",Body:TaskInfo{}})
			fmt.Println("I am taskgen ",t.GetTime().Sub(ZEROTIME).Milliseconds())
			
			return t
		}
		e.ApplyToAllComponent(CTaskGen, f)
	})

	frameNum := 12000
	for i := 0; i < frameNum; i++ {
		log.Println("simluating", i, "/", frameNum)
		simulator.Update()

	}


}

func NewCenterCluster() []Node {
	var nodes []Node
	nodes = append(nodes, Node{
		"user1",
		[]NodeComponent{
			CreateTaskGen("user1"),
		},
	})

	nodes = append(nodes, Node{
		"master1",
		[]NodeComponent{
			CreateScheduler("master1"),
		},
	})
	for i := 0; i < 10; i++ {

		workerName := fmt.Sprintf("worker%d", i)
		nodes = append(nodes, Node{
			workerName,
			[]NodeComponent{
				CreateResourceManager(workerName),
			},
		})

	}

	return nodes
}
