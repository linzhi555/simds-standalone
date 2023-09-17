package main

import (
	"flag"
	"fmt"
)

func NewClusterSimulator() *ECS {

	simulator := NewEcs()

	newNet := NewNetWork(10)
	netcard1 := NewNetCard("worker1")
	netcard1.JoinNetWork(newNet)
	netcard2 := NewNetCard("worker2")
	netcard2.JoinNetWork(newNet)

	simulator.AddEntities("network1")
	newNet.Waittings["worker1"] = &Message{From: "worker1", To: "worker2", Content: "newtask cpu1 ram2", LeftTime: 5}
	simulator.AddComponetsToEntity("network1", newNet)

	simulator.AddEntities("worker1")
	simulator.AddComponetsToEntity("worker1", &SystemTime{MicroSecond: 0}, netcard1)
	simulator.AddEntities("worker2")
	simulator.AddComponetsToEntity("worker2", &SystemTime{MicroSecond: 0}, netcard2)
	AddAllsystemToEcs(simulator)
	return simulator

}

func InitCenterSimulator() *ECS {
	simulator := NewEcs()

	// init network
	newNet := NewNetWork(100)
	simulator.AddEntities("network1", &SystemTime{MicroSecond: 0}, newNet)

	// init master
	newScheduler := NewScheduler("master1")
	newScheduler.Net.JoinNetWork(newNet)
	simulator.AddEntities("master1", &SystemTime{MicroSecond: 0}, newScheduler)

	// init taskGen
	newTaskgen := NewTaskGen("user1")
	newTaskgen.Net.JoinNetWork(newNet)
	simulator.AddEntities("user1", &SystemTime{MicroSecond: 0}, newTaskgen)

	// init taskGen
	const WorkerNum = 10
	for i := 0; i < WorkerNum; i++ {
		workerName := fmt.Sprintf("worker%d", i)
		newResourceManager := NewResourceManager(workerName)
		nodeinfo := NodeInfo{10, 10, 0, 0}

		nodeCopy := nodeinfo
		newScheduler.Workers[newResourceManager.Net.Addr] = &nodeCopy
		newResourceManager.Net.JoinNetWork(newNet)
		simulator.AddEntities(workerName, &SystemTime{MicroSecond: 0}, newResourceManager, &nodeinfo)
	}

	AddAllsystemToEcs(simulator)
	return simulator

}

var Debug = flag.Bool("debug", false, "run as debug mode")

func init() {
	flag.Parse()
}

func main() {
	//s := NewClusterSimulator()
	s := InitCenterSimulator()

	fmt.Println(s.Entities)
	for i := 0; i < 10000; i++ {
		s.Update()
		if *Debug && i%120 == 0 {
			fmt.Println("*************************")
			fmt.Println("*************************")
			fmt.Println(s)
			fmt.Println("*************************")
			fmt.Println("*************************")

		}
	}

}
