package main

import (
	"errors"
	"fmt"
)

type SystemTime struct {
	MicroSecond int32
}

func (s *SystemTime) Component() string {
	return "SystemTime"
}

type TaskInfo struct {
	Id            string
	CpuRequest    int32
	MemoryRequest int32
	SubmitTime    int32
	InQueneTime   int32
	StartTime     int32
	LifeTime      int32
	Status        string
}

func (t *TaskInfo) Component() string {
	return "TaskInfo"

}

type NodeInfo struct {
	Cpu            int32
	Memory         int32
	CpuAllocted    int32
	MemoryAllocted int32
}

func (n *NodeInfo) Component() string {
	return "NodeInfo"
}

type TaskList struct {
	AllTasks []*TaskInfo
}

func (n *TaskList) Component() string {
	return "TaskList"
}

type Message struct {
	From     string
	To       string
	Content  string
	LeftTime int32
	Body     interface{}
}

type MessageQueue struct {
	buffers []*Message
}

func NewMessageQueue() *MessageQueue {
	return &MessageQueue{
		buffers: make([]*Message, 0),
	}
}

func (mch *MessageQueue) Empty() bool {
	return mch.Len() == 0

}
func (mch *MessageQueue) Len() int {
	return len(mch.buffers)
}

func (mch *MessageQueue) InQueue(m *Message) {
	mch.buffers = append(mch.buffers, m)
}

func (mch *MessageQueue) Dequeue() (*Message, error) {
	if mch.Empty() == true {
		return nil, errors.New("the queue is Empty")
	}
	res := mch.buffers[0]
	mch.buffers = mch.buffers[1:mch.Len()]
	return res, nil
}

type Network struct {
	NetLatency int32
	Waittings  map[string]*Message
	Ins        map[string]*MessageQueue
	Outs       map[string]*MessageQueue
}

func (n *Network) Component() string {
	return "Network"
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

func (nc *NetCard) Component() string {
	return "NetCard"
}

type TaskGen struct {
	Net *NetCard
}

func NewTaskGen(hostname string) *TaskGen {
	return &TaskGen{
		Net: NewNetCard(hostname + ":" + "TaskGen"),
	}
}
func (t *TaskGen) Component() string {
	return "TaskGen"
}

func (nc *NetCard) JoinNetWork(net *Network) {
	nc.In = NewMessageQueue()
	nc.Out = NewMessageQueue()
	net.Outs[nc.Addr] = nc.In
	net.Ins[nc.Addr] = nc.Out
}

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
func (t *Scheduler) Component() string {
	return "Scheduler"
}

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
func (t *ResourceManager) Component() string {
	return "ResourceManager"
}
