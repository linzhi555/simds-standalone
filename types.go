package main

import "errors"

type MessageBody interface {
	MessageBody()
}

type TaskInfo struct {
	Id                string //the task id,it is unique
	CpuRequest        int32
	MemoryRequest     int32
	SubmitTime        int32
	InQueneTime       int32
	StartTime         int32
	LifeTime          int32
	Status            string
	Worker            string
	ScheduleFailCount int32
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

func (n *NodeInfo) Clone() *NodeInfo {
	var NodeInfoCopy = *n
	return &NodeInfoCopy
}

func (n *NodeInfo) AddAllocated(taskCpu, taskMemory int32) {
	n.CpuAllocted += taskCpu
	n.MemoryAllocted += taskMemory
}

func (n *NodeInfo) SubAllocated(taskCpu, taskMemory int32) {
	n.CpuAllocted -= taskCpu
	n.MemoryAllocted -= taskMemory
}

func (n *NodeInfo) CanAllocate(taskCpu, taskMemory int32) bool {
	if n.Cpu-n.CpuAllocted >= taskCpu && n.Memory-n.MemoryAllocted >= taskMemory {
		return true
	} else {
		return false
	}
}

func (TaskInfo) MessageBody() {}
func (NodeInfo) MessageBody() {}

type Message struct {
	From     string
	To       string
	Content  string
	LeftTime int32
	Body     MessageBody
}

type MessageQueue struct {
	buffers []Message
}

func NewMessageQueue() *MessageQueue {
	return &MessageQueue{
		buffers: make([]Message, 0),
	}
}

func (mch *MessageQueue) Empty() bool {
	return mch.Len() == 0

}
func (mch *MessageQueue) Len() int {
	return len(mch.buffers)
}

func (mch *MessageQueue) InQueue(m Message) {
	mch.buffers = append(mch.buffers, m)
}

func (mch *MessageQueue) Dequeue() (Message, error) {
	if mch.Empty() == true {
		return Message{}, errors.New("the queue is Empty")
	}
	res := mch.buffers[0]
	mch.buffers = mch.buffers[1:mch.Len()]
	return res, nil
}
