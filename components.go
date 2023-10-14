package main

import (
	"fmt"
	"time"
)

const MiliSecond int32 = 1
const Second int32 = 1000

const CMockNetWork ComponentName = "MockNetwork"
const CTaskGen ComponentName = "TaskGen"
const CScheduler ComponentName = "Scheduler"
const CResouceManger ComponentName = "ResourceManager"
const CStateStorage ComponentName = "StateStorage"

type OsApi interface {
	GetTime() time.Time
	Net() NetInterface
}

type NodeComponent interface {
	Component
	SetOsApi(OsApi)
}

type MockNetwork struct {
	Os         OsApi
	NetLatency int32
	Waittings  Vec[Message]
	Ins        map[string]*Vec[Message]
	Outs       map[string]*Vec[Message]
}

func NewMockNetWork(latency int32) *MockNetwork {
	return &MockNetwork{
		NetLatency: latency,
		Waittings:  Vec[Message]{},
		Ins:        make(map[string]*Vec[Message]),
		Outs:       make(map[string]*Vec[Message]),
	}
}

func (n MockNetwork) Component() ComponentName { return CMockNetWork }
func (n *MockNetwork) SetOsApi(osapi OsApi)    { n.Os = osapi }

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
	Os        OsApi
	Host      string
	StartTime time.Time
	CurTaskId int
	Receivers []string
}

func NewTaskGen(hostname string) *TaskGen {
	return &TaskGen{
		Host:      hostname,
		CurTaskId: 0,
	}
}

func (n TaskGen) Component() ComponentName { return CTaskGen }
func (n *TaskGen) SetOsApi(osapi OsApi)    { n.Os = osapi }

type Scheduler struct {
	Os           OsApi
	Host         string
	Workers      map[string]*NodeInfo
	LocalNode    *NodeInfo
	WaitSchedule Vec[TaskInfo]
	TasksStatus  map[string]*TaskInfo
}

func NewScheduler(hostname string) *Scheduler {
	return &Scheduler{
		Host:         hostname,
		Workers:      make(map[string]*NodeInfo),
		WaitSchedule: Vec[TaskInfo]{},
		TasksStatus:  make(map[string]*TaskInfo),
	}
}

func (n Scheduler) Component() ComponentName { return CScheduler }
func (n *Scheduler) SetOsApi(osapi OsApi)    { n.Os = osapi }

func (s *Scheduler) GetAllWokersName() []string {
	keys := make([]string, 0, len(s.Workers))
	for k := range s.Workers {
		keys = append(keys, k)
	}
	return keys
}

type StateStorage struct {
	Os           OsApi
	Host         string
	StartTime 	 time.Time
	Workers      map[string]*NodeInfo
}

func NewStateStorage(hostname string) *StateStorage {
	return &StateStorage{
		Host:         hostname,
		Workers:      make(map[string]*NodeInfo),
	}
}

func (s  *StateStorage)StateCopy()Vec[NodeInfo]{
	nodes := make(Vec[NodeInfo],0,len(s.Workers))
	for _, ni := range s.Workers {
		nodes = append(nodes, *ni)
	}
	return nodes
}


func (s StateStorage) Component() ComponentName { return CStateStorage }
func (s *StateStorage) SetOsApi(osapi OsApi)    { s.Os = osapi }



type ResourceManager struct {
	Os                 OsApi
	Host               string
	Tasks              map[string]*TaskInfo
	Node               NodeInfo
	TaskFinishReceiver string // if it is not zero , the receiver wiil get the notifiction
}

func NewResourceManager(host string) *ResourceManager {
	return &ResourceManager{
		Host:  host,
		Tasks: make(map[string]*TaskInfo),
	}
}

func (n ResourceManager) Component() ComponentName { return CResouceManger }
func (n *ResourceManager) SetOsApi(osapi OsApi)    { n.Os = osapi }
