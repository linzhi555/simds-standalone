package main

import (
	"fmt"
)

const MiliSecond int32 = 1
const Second int32 = 1000

// SystemTime Component, a entity can know it when it has this obecjt
const CSystemTime ComponentName = "SystemTime"

type SystemTime struct {
	Time int32
}

func (s SystemTime) Component() ComponentName {
	return CSystemTime
}

const CNetWork ComponentName = "Network"

type Network struct {
	NetLatency int32
	Waittings  Vec[Message]
	Ins        map[string]*Vec[Message]
	Outs       map[string]*Vec[Message]
}

func (n Network) Component() ComponentName {
	return CNetWork
}

func CreateNetWork(latency int32) Network {
	return Network{
		NetLatency: latency,
		Waittings:  Vec[Message]{},
		Ins:        make(map[string]*Vec[Message]),
		Outs:       make(map[string]*Vec[Message]),
	}
}

func (n Network) String() string {
	var res string
	res += "Waittings: \n"
	for _, v := range n.Waittings {
		res += fmt.Sprintln(v)
	}
	res += "Routes: \n"
	for k := range n.Outs {
		res += fmt.Sprintln(k)
	}

	return res
}

const CNetCard = "NetCard"

type NetCard struct {
	Addr string
	In   *Vec[Message]
	Out  *Vec[Message]
}

func CreateNetCard(name string) NetCard {
	return NetCard{
		Addr: name,
	}

}

func (nc NetCard) Component() ComponentName {
	return CNetCard
}

const CTaskGen = "NetCard"

type TaskGen struct {
	CurTaskId int
	Net       NetCard
	Receivers []string
}

func CreateTaskGen(hostname string) TaskGen {
	return TaskGen{
		CurTaskId: 0,
		Net:       CreateNetCard(hostname + ":" + "TaskGen"),
	}
}
func (t TaskGen) Component() ComponentName {
	return CTaskGen
}

func (nc *NetCard) JoinNetWork(net *Network) {
	nc.In = &Vec[Message]{}
	nc.Out = &Vec[Message]{}
	net.Outs[nc.Addr] = nc.In
	net.Ins[nc.Addr] = nc.Out
}

const CScheduler ComponentName = "Scheduler"

type Scheduler struct {
	Net     NetCard
	Workers map[string]*NodeInfo
	Tasks   map[string]*Vec[TaskInfo]
}

func CreateScheduler(hostname string) Scheduler {
	return Scheduler{
		Net:     CreateNetCard(hostname + ":" + "Scheduler"),
		Workers: make(map[string]*NodeInfo),
		Tasks:   make(map[string]*Vec[TaskInfo]),
	}
}
func (s Scheduler) Component() ComponentName {
	return CScheduler
}

func (s *Scheduler) GetAllWokersName() []string {
	keys := make([]string, 0, len(s.Workers))
	for k := range s.Workers {
		keys = append(keys, k)
	}
	return keys
}

const CResouceManger ComponentName = "ResourceManager"

type ResourceManager struct {
	Tasks              map[string]*TaskInfo
	Net                NetCard
	Node               NodeInfo
	TaskFinishReceiver string // if it is not zero , the receiver wiil get the notifiction
}

func CreateResourceManager(host string) ResourceManager {
	return ResourceManager{
		Tasks: make(map[string]*TaskInfo),
		Net:   CreateNetCard(host + ":" + "ResourceManager"),
	}
}
func (t ResourceManager) Component() ComponentName {
	return CResouceManger
}
