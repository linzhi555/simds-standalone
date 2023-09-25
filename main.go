package main

import (
	"flag"
	"fmt"
	"math/rand"
)

var Debug = flag.Bool("debug", false, "run as debug mode")
var Dcss = flag.Bool("dcss", false, "run dcss")

func init() {
	flag.Parse()
}

func InitCenterSimulator() *ECS {
	simulator := NewEcs()

	// init network
	newNet := NewNetWork(Config.NetLatency * MiliSecond)
	simulator.AddEntities("network1", &SystemTime{Time: 0}, newNet)

	// init master
	newScheduler := NewScheduler("master0")
	newScheduler.Net.JoinNetWork(newNet)
	simulator.AddEntities("master0", &SystemTime{Time: 0}, newScheduler)

	// init taskGen
	newTaskgen := NewTaskGen("user1")
	newTaskgen.Net.JoinNetWork(newNet)
	simulator.AddEntities("user1", &SystemTime{Time: 0}, newTaskgen)

	// init nodes
	var WorkerNum = Config.NodeNum
	for i := 0; i < int(WorkerNum); i++ {
		workerName := fmt.Sprintf("worker%d", i)
		newResourceManager := NewResourceManager(workerName)
		nodeinfo := &NodeInfo{Config.NodeCpu, Config.NodeMemory, 0, 0}
		newScheduler.Workers[newResourceManager.Net.Addr] = nodeinfo.Clone()
		newResourceManager.Net.JoinNetWork(newNet)
		newResourceManager.Node = nodeinfo.Clone()
		newResourceManager.TaskFinishReceiver = newScheduler.Net.Addr

		simulator.AddEntities(EntityName(workerName), &SystemTime{Time: 0}, newResourceManager)
	}

	RegisteCentralizedsystemToEcs(simulator)
	return simulator

}

func InitDcssSimulator() *ECS {
	simulator := NewEcs()

	// init network
	newNet := NewNetWork(Config.NetLatency * MiliSecond)
	simulator.AddEntities("network1", &SystemTime{Time: 0}, newNet)

	// init taskGen
	newTaskgen := NewTaskGen("user1")
	newTaskgen.Net.JoinNetWork(newNet)
	simulator.AddEntities("user1", &SystemTime{Time: 0}, newTaskgen)

	// init nodes these nodes are scheduler and worker in same time.
	var nodeNum = int(Config.NodeNum)
	for i := 0; i < nodeNum; i++ {
		nodeName := fmt.Sprintf("node%d", i)
		newResourceManager := NewResourceManager(nodeName)
		newScheduler := NewScheduler(nodeName)
		initNeiborhood(newScheduler, nodeNum, int(Config.DcssNeibor))
		newResourceManager.Net.JoinNetWork(newNet)
		newResourceManager.Node = &NodeInfo{Config.NodeCpu, Config.NodeMemory, 0, 0}
		simulator.AddEntities(EntityName(nodeName), &SystemTime{Time: 0}, newResourceManager, newScheduler)
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
		scheduler.Workers[n] = &NodeInfo{Config.NodeCpu, Config.NodeMemory, 0, 0}
	}
}

func main() {
	startPerf()
	defer stopPerf()
	//s := NewClusterSimulator()
	var s *ECS
	if *Dcss {
		s = InitDcssSimulator()
	} else {
		s = InitCenterSimulator()
	}
	fmt.Println(s.Entities)
	for i := 0; i < 12000; i++ {
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
