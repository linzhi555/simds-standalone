package main

import "fmt"

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

	simulator.AddSystem("SystemTime", func(c Component) {
		t := c.(*SystemTime)
		t.MicroSecond += 1
	})

	simulator.AddSystem("NetCard", NetCardTicks)

	simulator.AddSystem("Network", NetworkUpdate)
	return simulator
}

const Debug = false

func main() {
	s := NewClusterSimulator()

	for i := 0; i < 10; i++ {
		s.Ticks()
		if Debug {

			fmt.Println("vvvvvvvvvvvvvvvv")
			fmt.Println(s)
			fmt.Println("^^^^^^^^^^^^^^^^")

		}
	}

}
