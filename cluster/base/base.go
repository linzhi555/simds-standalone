package base

// types.go 定义一系列的基础类型

import (
	"encoding/json"
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
}

func (b *BasicActor) GetAddress() string {
	return b.Host
}

func (b *BasicActor) SetOsApi(os OsApi) {
	b.Os = os
}


type Actor interface {
	GetAddress() string
	Debug()
	SetOsApi(OsApi)
	Update(Message)

	// below is only  for simulation mode not for deploy mode
	SimulateTasksUpdate()
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
