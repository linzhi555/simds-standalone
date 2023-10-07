package main

import "errors"


type NetInterface interface{
	Empty()bool
	Recv()(Message,error)
	Send(Message)error
}

type MockNetCard struct {
	Addr string
	In   *Vec[Message]
	Out  *Vec[Message]
}

func CreateMockNetCard(name string) MockNetCard {
	return MockNetCard{
		Addr: name,
	}
}

func (card MockNetCard)Empty()bool{

	return true
}

func (card MockNetCard)Recv()(Message,error){
	return Message{},errors.New("recv fail")
}

func (card MockNetCard)Send(m Message)error{
	return errors.New("send fail")
}

func (nc *MockNetCard) JoinNetWork(net *MockNetwork) {
	nc.In = &Vec[Message]{}
	nc.Out = &Vec[Message]{}
	net.Outs[nc.Addr] = nc.In
	net.Ins[nc.Addr] = nc.Out
}

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

type Vec[T TaskInfo | NodeInfo | Message] []T

func (vec *Vec[T])InQueueFront(data T){
	*vec = append(Vec[T]{data},*vec...)

}

func (vec *Vec[T]) InQueue(data T) {
	*vec = append(*vec, data)
}

func (vec *Vec[T]) Len() int {
	return len(*vec)
}
func (vec *Vec[T]) Empty() bool {
	return vec.Len() == 0
}
func (vec *Vec[T]) Dequeue() (T, error) {
	var res T
	if vec.Empty() == true {
		return res, errors.New("the queue is Empty")
	}
	res = (*vec)[0]
	*vec = (*vec)[1:vec.Len()]
	return res, nil
}

func (vec *Vec[T]) Delete(index int) {
	*vec = append((*vec)[0:index], (*vec)[index+1:vec.Len()]...)

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

//type MessageQueue Vec[Message]
//func NewMessageQueue() *MessageQueue{
//	return &MessageQueue{}
//}

//type MessageQueue struct {
//	buffers []Message
//}
//
//func NewMessageQueue() *MessageQueue {
//	return &MessageQueue{
//		buffers: make([]Message, 0),
//	}
//}
//
//func (mch *MessageQueue) Empty() bool {
//	return mch.Len() == 0
//
//}
//func (mch *MessageQueue) Len() int {
//	return len(mch.buffers)
//}
//
//func (mch *MessageQueue) InQueue(m Message) {
//	mch.buffers = append(mch.buffers, m)
//}
//
//func (mch *MessageQueue) Dequeue() (Message, error) {
//	if mch.Empty() == true {
//		return Message{}, errors.New("the queue is Empty")
//	}
//	res := mch.buffers[0]
//	mch.buffers = mch.buffers[1:mch.Len()]
//	return res, nil
//}
