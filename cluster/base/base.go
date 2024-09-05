package base

// types.go 定义一系列的基础类型

import (
	"encoding/json"
	"errors"
	"time"

	"simds-standalone/common"
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

var InverseJsonTable map[string]func(string) MessageBody = map[string]func(string) MessageBody{}

func FromJson(head string, body string) MessageBody {

	for pattern, f := range InverseJsonTable {
		if common.MatchPattern(pattern, head) {
			return f(body)
		}
	}

	panic("wrong type contentType")
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
func (vec *Vec[T]) InQueueBack(data T) {
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
	if index >= vec.Len() || index < 0 {
		panic("index out of range")
	} else if index == vec.Len()-1 {
		*vec = (*vec)[:vec.Len()-1]
	} else {
		copy((*vec)[index:], (*vec)[index+1:])
		*vec = (*vec)[:vec.Len()-1]
	}
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
	Head     string
	Body     MessageBody
	LeftTime time.Duration
}

// NetInterface 用于处理 Message
type NetInterface interface {
	Send(Message) error
}
