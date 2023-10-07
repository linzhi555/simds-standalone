package main

import (
	"fmt"
	"simds-standalone/ecs"
	"time"
)

const MiliSecond int32 = 1
const Second int32 = 1000

const CMockNetWork ecs.ComponentName = "MockNetwork"
const CTaskGen ecs.ComponentName = "TaskGen"
const CScheduler ecs.ComponentName = "Scheduler"
const CResouceManger ecs.ComponentName = "ResourceManager"

type NodeComponent interface {
	ecs.Component
	SetComponent(n ecs.ComponentName)
	InitNet(NetInterface)
	InitTimeGetter(func() time.Time)
	Net() NetInterface
	GetTime() time.Time
}

type baseComp struct {
	name        ecs.ComponentName
	getTimeFunc func() time.Time
	net         NetInterface
}

func (n *baseComp) Component() ecs.ComponentName {
	return n.name
}

func (n *baseComp) SetComponent(name ecs.ComponentName) {
	n.name = name
}

func (n *baseComp) InitNet(ni NetInterface) {
	n.net = ni
}

func (n *baseComp) InitTimeGetter(f func() time.Time) {
	n.getTimeFunc = f
}

func (n *baseComp) Net() NetInterface {
	return n.net
}
func (n *baseComp) GetTime() time.Time {
	return n.getTimeFunc()
}

type MockNetwork struct {
	*baseComp
	NetLatency int32
	Waittings  Vec[Message]
	Ins        map[string]*Vec[Message]
	Outs       map[string]*Vec[Message]
}

func CreateMockNetWork(latency int32) MockNetwork {
	return MockNetwork{
		baseComp:   &baseComp{name: CMockNetWork},
		NetLatency: latency,
		Waittings:  Vec[Message]{},
		Ins:        make(map[string]*Vec[Message]),
		Outs:       make(map[string]*Vec[Message]),
	}
}

func (n MockNetwork) String() string {
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

type TaskGen struct {
	*baseComp
	CurTaskId int
	Receivers []string
}

func CreateTaskGen(hostname string) TaskGen {
	return TaskGen{
		baseComp:  &baseComp{name: CTaskGen},
		CurTaskId: 0,
	}
}

type Scheduler struct {
	*baseComp
	Workers      map[string]*NodeInfo
	WaitSchedule Vec[TaskInfo]
	TasksStatus  map[string]*TaskInfo
}

func CreateScheduler(hostname string) Scheduler {
	return Scheduler{
		baseComp:     &baseComp{name: CScheduler},
		Workers:      make(map[string]*NodeInfo),
		WaitSchedule: Vec[TaskInfo]{},
		TasksStatus:  make(map[string]*TaskInfo),
	}
}

func (s *Scheduler) GetAllWokersName() []string {
	keys := make([]string, 0, len(s.Workers))
	for k := range s.Workers {
		keys = append(keys, k)
	}
	return keys
}

type ResourceManager struct {
	*baseComp
	Tasks              map[string]*TaskInfo
	Node               NodeInfo
	TaskFinishReceiver string // if it is not zero , the receiver wiil get the notifiction
}

func CreateResourceManager(host string) ResourceManager {
	return ResourceManager{
		baseComp: &baseComp{name: CResouceManger},
		Tasks:    make(map[string]*TaskInfo),
	}
}
