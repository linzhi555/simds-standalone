package main

import (
	"flag"
	"fmt"
)

func NewClusterSimulator() *ECS {

	simulator := NewEcs()

	newNet := NewNetWork()
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
	newNet := NewNetWork()
	newScheduler := NewScheduler("master1")
	newTaskgen := NewTaskGen("user1")
	newScheduler.Net.JoinNetWork(newNet)
	newTaskgen.Net.JoinNetWork(newNet)

	simulator.AddEntities("network1", &SystemTime{MicroSecond: 0}, newNet)
	simulator.AddEntities("master1", &SystemTime{MicroSecond: 0}, newScheduler)
	simulator.AddEntities("user1", &SystemTime{MicroSecond: 0}, newTaskgen)

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

	for i := 0; i < 20; i++ {
		s.Update()
		if *Debug {
			fmt.Println("*************************")
			fmt.Println("*************************")
			fmt.Println(s)
			fmt.Println("*************************")
			fmt.Println("*************************")

		}
	}

}
