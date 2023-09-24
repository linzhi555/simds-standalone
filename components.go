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

func (s *SystemTime) Component() ComponentName {
	return CSystemTime
}

const CTaskInfo ComponentName = "TaskInfo"

type TaskInfo struct {
	Id            string //the task id,it is unique
	CpuRequest    int32
	MemoryRequest int32
	SubmitTime    int32
	InQueneTime   int32
	StartTime     int32
	LifeTime      int32
	Status        string
	Worker        string
}

func (t *TaskInfo) Component() ComponentName {
	return CTaskInfo

}

func (t *TaskInfo) DeepCopy() *TaskInfo {
	var newT TaskInfo = *t
	return &newT
}

const CNodeInfo ComponentName = "NodeInfo"

type NodeInfo struct {
	Cpu            int32
	Memory         int32
	CpuAllocted    int32
	MemoryAllocted int32
}

func (n *NodeInfo) Component() ComponentName {
	return CNodeInfo
}

func (n *NodeInfo) CanAllocate(taskCpu, taskMemory int32) bool {
	if n.Cpu-n.CpuAllocted >= taskCpu && n.Memory-n.MemoryAllocted >= taskMemory {
		return true
	} else {
		return false
	}
}

const CNetWork ComponentName = "Network"

type Network struct {
	NetLatency int32
	Waittings  map[string]*Message
	Ins        map[string]*MessageQueue
	Outs       map[string]*MessageQueue
}

func (n *Network) Component() ComponentName {
	return CNetWork
}

func NewNetWork(latency int32) *Network {
	return &Network{
		NetLatency: latency,
		Waittings:  make(map[string]*Message),
		Ins:        make(map[string]*MessageQueue),
		Outs:       make(map[string]*MessageQueue),
	}
}

func (n *Network) String() string {
	var res string
	res += "Waittings: \n"
	for _, v := range n.Waittings {
		if v == nil {
			continue
		}
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
	In   *MessageQueue
	Out  *MessageQueue
}

func NewNetCard(name string) *NetCard {
	return &NetCard{
		Addr: name,
	}

}
func (nc *NetCard) Component() ComponentName {
	return CNetCard
}

const CTaskGen = "NetCard"

type TaskGen struct {
	CurTaskId int
	Net       *NetCard
	Receivers []string
}

func NewTaskGen(hostname string) *TaskGen {
	return &TaskGen{
		CurTaskId: 0,
		Net:       NewNetCard(hostname + ":" + "TaskGen"),
	}
}
func (t *TaskGen) Component() ComponentName {
	return CTaskGen
}

func (nc *NetCard) JoinNetWork(net *Network) {
	nc.In = NewMessageQueue()
	nc.Out = NewMessageQueue()
	net.Outs[nc.Addr] = nc.In
	net.Ins[nc.Addr] = nc.Out
}

const CScheduler ComponentName = "Scheduler"

type Scheduler struct {
	Net     *NetCard
	Workers map[string]*NodeInfo
	Tasks   map[string]*TaskInfo
}

func NewScheduler(hostname string) *Scheduler {
	return &Scheduler{
		Net:     NewNetCard(hostname + ":" + "Scheduler"),
		Workers: make(map[string]*NodeInfo),
		Tasks:   make(map[string]*TaskInfo),
	}
}
func (t *Scheduler) Component() ComponentName {
	return CScheduler
}

const CResouceManger ComponentName = "ResourceManager"

type ResourceManager struct {
	Tasks map[string]*TaskInfo
	Net   *NetCard
}

func NewResourceManager(host string) *ResourceManager {
	return &ResourceManager{
		Tasks: make(map[string]*TaskInfo),
		Net:   NewNetCard(host + ":" + "ResourceManager"),
	}
}
func (t *ResourceManager) Component() ComponentName {
	return CResouceManger
}
