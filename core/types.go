package core

// types.go 定义一系列的基础类型

import (
	"errors"
	"time"
)

// MessageBody Message的Body字段
type MessageBody interface {
	MessageBody()
}

// TaskInfo 任务的基本信息，还有一些附加的调度器使用的字段
type TaskInfo struct {
	Id                string //the task id,it is unique
	CpuRequest        int32
	MemoryRequest     int32
	StartTime         time.Time
	LifeTime          time.Duration
	Status            string
	Worker            string
	ScheduleFailCount int32
}

// MessageBody TaskInfo 是MessageBody
func (TaskInfo) MessageBody() {}

// Clone 复制新的TaskInfo
func (t *TaskInfo) Clone() *TaskInfo {
	var newT TaskInfo = *t
	return &newT
}

// NodeInfo  节点资源信息
type NodeInfo struct {
	Addr           string
	Cpu            int32
	Memory         int32
	CpuAllocted    int32
	MemoryAllocted int32
}

// MessageBody NodeInfo 是MessageBody
func (NodeInfo) MessageBody() {}

// Clone 复制新的NodeInfo
func (n *NodeInfo) Clone() *NodeInfo {
	var NodeInfoCopy = *n
	return &NodeInfoCopy
}

// AddAllocated 更新节点信息-增加已分配
func (n *NodeInfo) AddAllocated(taskCpu, taskMemory int32) {
	n.CpuAllocted += taskCpu
	n.MemoryAllocted += taskMemory
}

// SubAllocated 更新节点信息-释放已分配
func (n *NodeInfo) SubAllocated(taskCpu, taskMemory int32) {
	n.CpuAllocted -= taskCpu
	n.MemoryAllocted -= taskMemory
}

// CanAllocate 判读是否满足分配
func (n *NodeInfo) CanAllocate(taskCpu, taskMemory int32) bool {
	if n.Cpu-n.CpuAllocted >= taskCpu && n.Memory-n.MemoryAllocted >= taskMemory {
		return true
	}
	return false
}

// CanAllocateTask 判读是否满足分配某个任务
func (n *NodeInfo) CanAllocateTask(task *TaskInfo) bool {
	return n.CanAllocate(task.CpuRequest, task.MemoryRequest)
}

// Vec 为三种类型定义Vector
type Vec[T TaskInfo | NodeInfo | Message] []T

// MessageBody Vec[T] 是 MessageBody
func (vec Vec[T]) MessageBody() {}

// InQueueFront 在Vector头部入队
func (vec *Vec[T]) InQueueFront(data T) {
	*vec = append(Vec[T]{data}, *vec...)

}

// Clone 拷贝一份新的Vec[T]
func (vec *Vec[T]) Clone() *Vec[T] {
	newVec := make(Vec[T], len(*vec))
	for i, data := range *vec {
		newVec[i] = data
	}
	return &newVec
}

// InQueue 在Vector尾部入队
func (vec *Vec[T]) InQueue(data T) {
	*vec = append(*vec, data)
}

// Len 返回Vector 长度
func (vec *Vec[T]) Len() int {
	return len(*vec)
}

// Empty 返回Vector 是否为空
func (vec *Vec[T]) Empty() bool {
	return vec.Len() == 0
}

// Dequeue 在Vector头部出队
func (vec *Vec[T]) Dequeue() (T, error) {
	var res T
	if vec.Empty() == true {
		return res, errors.New("the queue is Empty")
	}
	res = (*vec)[0]
	*vec = (*vec)[1:vec.Len()]
	return res, nil
}

// Delete 删除Vector的某个元素
func (vec *Vec[T]) Delete(index int) {
	*vec = append((*vec)[0:index], (*vec)[index+1:vec.Len()]...)

}

// Message 用于组件信息传递
type Message struct {
	From     string
	To       string
	Content  string
	LeftTime int32
	Body     MessageBody
}

// NetInterface 用于处理 Message
type NetInterface interface {
	Empty() bool
	Recv() (Message, error)
	Send(Message) error
	GetAddr() string
}