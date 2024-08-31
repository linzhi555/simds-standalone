package base

// types.go 定义一系列的基础类型

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Cluster struct {
	Nodes []Node
}

func (c *Cluster) Join(node Node) {
	c.Nodes = append(c.Nodes, node)
}

type Node struct {
	Actors []Actor
}

func NewNode(actors ...Actor) Node {
	var node Node
	for _, actor := range actors {
		node.Actors = append(node.Actors, actor)
	}
	return node
}

type BasicActor struct {
	Os             OsApi
	Host           string
	NextUpdateTime time.Time
}

func (b *BasicActor) GetHostName() string {
	return b.Host
}

func (b *BasicActor) SetOsApi(os OsApi) {
	b.Os = os
}

func (b *BasicActor) GetNextUpdateTime() time.Time {
	return b.NextUpdateTime
}

func (b *BasicActor) SetNextUpdateTime(t time.Time) {
	b.NextUpdateTime = t
}

type Actor interface {
	GetHostName() string
	Debug()
	SetOsApi(OsApi)
	Update(Message)

	// below is only  for simulation mode not for deploy mode
	SimulateTasksUpdate()
	GetNextUpdateTime() time.Time
	SetNextUpdateTime(t time.Time)
}

// OsApi 系统调用 抽象接口
type OsApi interface {
	NetInterface
	GetTime() time.Time
	Run(f func())
}

// MessageBody Message的Body字段
type MessageBody interface {
	MessageBody()
}

func ToJson[T MessageBody](t T) string {
	bytes, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func FromJson(contentType string, s string) MessageBody {
	if strings.HasPrefix(contentType, "Task") {
		var res TaskInfo
		err := json.Unmarshal([]byte(s), &res)
		if err != nil {
			panic(err)
		}
		return res
	} else if strings.HasPrefix(contentType, "NodeInfos") {
		var res Vec[NodeInfo]
		err := json.Unmarshal([]byte(s), &res)
		if err != nil {
			panic(err)
		}
		return res
	} else if strings.HasPrefix(contentType, "Signal") {
		return Signal(s)
	} else {
		panic("wrong type contentType")
	}
}

type Signal string

func (s Signal) MessageBody() {}

// TaskInfo 任务的基本信息，还有一些附加的调度器使用的字段
type TaskInfo struct {
	Id                string //the task id,it is unique
	CpuRequest        int32
	MemoryRequest     int32
	StartTime         time.Time
	LifeTime          time.Duration
	LeftTime          time.Duration
	Status            string
	User              string
	Worker            string
	ScheduleFailCount int32
	Cmd               string
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

func (n *NodeInfo) CpuPercent() float32 {
	return float32(n.CpuAllocted) / float32(n.Cpu)
}

func (n *NodeInfo) MemoryPercent() float32 {
	return float32(n.MemoryAllocted) / float32(n.Memory)
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
type Vec[T any] []T

// MessageBody Vec[T] 是 MessageBody
func (vec Vec[T]) MessageBody() {}

// InQueueFront 在Vector头部入队
func (vec *Vec[T]) InQueueFront(data T) {
	*vec = append(Vec[T]{data}, *vec...)

}

// Clone 拷贝一份新的Vec[T]
func (vec *Vec[T]) Clone() *Vec[T] {
	newVec := make(Vec[T], len(*vec))
	copy(newVec, *vec)
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
	if vec.Empty() {
		return res, errors.New("the queue is Empty")
	}
	res = (*vec)[0]
	*vec = (*vec)[1:vec.Len()]
	return res, nil
}

// Dequeue 在Vector尾部出队
func (vec *Vec[T]) Pop() (T, error) {
	var res T
	if vec.Empty() {
		return res, errors.New("the queue is Empty")
	}
	res = (*vec)[vec.Len()-1]
	*vec = (*vec)[0 : vec.Len()-1]
	return res, nil
}

// Delete 删除Vector的某个元素
func (vec *Vec[T]) Delete(index int) {
	*vec = append((*vec)[0:index], (*vec)[index+1:vec.Len()]...)
}

// 清空Vector
func (vec *Vec[T]) Clean() {
	*vec = (*vec)[0:0]
}

// Message 用于组件信息传递
type Message struct {
	Id       string
	From     string
	To       string
	Content  string
	LeftTime time.Duration
	Body     MessageBody
}

// NetInterface 用于处理 Message
type NetInterface interface {
	Send(Message) error
}
