package main

import (
	"flag"
	"fmt"
	"math/rand"
)

func InitCenterSimulator() *ECS {
	simulator := NewEcs()

	// init network
	newNet := NewNetWork(10 * MiliSecond)
	simulator.AddEntities("network1", &SystemTime{Time: 0}, newNet)

	// init master
	newScheduler := NewScheduler("master1")
	newScheduler.Net.JoinNetWork(newNet)
	simulator.AddEntities("master1", &SystemTime{Time: 0}, newScheduler)

	// init taskGen
	newTaskgen := NewTaskGen("user1")
	newTaskgen.Net.JoinNetWork(newNet)
	simulator.AddEntities("user1", &SystemTime{Time: 0}, newTaskgen)

	// init taskGen
	const WorkerNum = 20
	for i := 0; i < WorkerNum; i++ {
		workerName := fmt.Sprintf("worker%d", i)
		newResourceManager := NewResourceManager(workerName)
		nodeinfo := NodeInfo{10, 10, 0, 0}

		nodeCopy := nodeinfo
		newScheduler.Workers[newResourceManager.Net.Addr] = &nodeCopy
		newResourceManager.Net.JoinNetWork(newNet)
		simulator.AddEntities(EntityName(workerName), &SystemTime{Time: 0}, newResourceManager, &nodeinfo)
	}

	RegisteCentralizedsystemToEcs(simulator)
	return simulator

}

func InitDcssSimulator() *ECS {
	simulator := NewEcs()

	// init network
	newNet := NewNetWork(10 * MiliSecond)
	simulator.AddEntities("network1", &SystemTime{Time: 0}, newNet)

	// init taskGen
	newTaskgen := NewTaskGen("user1")
	newTaskgen.Net.JoinNetWork(newNet)
	simulator.AddEntities("user1", &SystemTime{Time: 0}, newTaskgen)

	// init nodes these nodes are scheduler and worker in same time.
	const nodeNum = 20
	for i := 0; i < nodeNum; i++ {
		nodeName := fmt.Sprintf("node%d", i)
		newResourceManager := NewResourceManager(nodeName)
		newScheduler := NewScheduler(nodeName)
		nodeinfo := NodeInfo{10, 10, 0, 0}
		newScheduler.Net.JoinNetWork(newNet)
		initNeiborhood(newScheduler, nodeNum, nodeNum/3)
		newResourceManager.Net.JoinNetWork(newNet)
		simulator.AddEntities(EntityName(nodeName), &SystemTime{Time: 0}, newResourceManager, newScheduler, &nodeinfo)
	}

	RegisteDcssSystemToEcs(simulator)
	return simulator

}

func initNeiborhood(scheduler *Scheduler, allNodeNum int, neiborNum int) {

	var neibors []string
	// add self in the first for convernience, and ignore the first when actually register neibor
	neibors = append(neibors, scheduler.Net.Addr)
	for len(neibors) != neiborNum+1 {
		newNeibor := fmt.Sprintf("node%d:Scheduler", rand.Intn(allNodeNum))
		alreadyExisted := false
		for _, n := range neibors {
			if n == newNeibor {
				alreadyExisted = true
			}
		}
		if !alreadyExisted {
			neibors = append(neibors, newNeibor)
		}
	}
	for _, n := range neibors[1:] {
		scheduler.Workers[n] = &NodeInfo{10, 10, 0, 0}
	}
}

var Debug = flag.Bool("debug", false, "run as debug mode")
var Dcss = flag.Bool("dcss", false, "run dcss")

func init() {
	flag.Parse()
}

func main() {
	//s := NewClusterSimulator()
	var s *ECS
	if *Dcss {
		s = InitDcssSimulator()
	} else {
		s = InitCenterSimulator()
	}

	fmt.Println(s.Entities)
	for i := 0; i < 10000; i++ {
		s.Update()
		if *Debug && i == 1000 {
			fmt.Println("*************************")
			fmt.Println("*************************")
			fmt.Println(s)

			fmt.Println("*************************")
			fmt.Println("*************************")

		}
	}

}
