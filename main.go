package main

import "fmt"


func NewClusterSimulator()*ECS{
	simulator := NewEcs()
	simulator.AddEntities("worker1")
	simulator.AddComponetsToEntity("worker1",&SystemTime{MicroSecond:0})
	simulator.AddEntities("worker2")
	simulator.AddComponetsToEntity("worker2",&SystemTime{MicroSecond:0})

	simulator.AddSystem("SystemTime",func(c Component) {
		t := c.(*SystemTime)
		t.MicroSecond += 1
	})
	return simulator
}

func main(){
	s := NewClusterSimulator()

	for i:=0 ;i<100;i++ {
		s.Ticks()
	}

	fmt.Println(s)




}
