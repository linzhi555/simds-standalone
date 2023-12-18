package main

//ecs_runner.go 使用 通用的ecs.go定义的通用ECS来模拟XXXCluster

import (
	"fmt"
	"log"
	"simds-standalone/common"
	"time"
)

// ZEROTIME 模拟开始的现实时间，以此作为模拟器的零点时间
var ZEROTIME time.Time = time.Now()

// MockOs 为组件提供模拟的系统调用
type MockOs struct {
	getTimeFunc func() time.Time
	net         NetInterface
}

// GetTime 提供模拟时间
func (o *MockOs) GetTime() time.Time {
	return o.getTimeFunc()
}

// Net 提供模拟的网络接口
func (o *MockOs) Net() NetInterface {
	return o.net
}

// MockNetCard 模拟的网卡
type MockNetCard struct {
	Addr string
	In   *Vec[Message]
	Out  *Vec[Message]
}

// CreateMockNetCard 创建一个未连接网络的模拟网卡
func CreateMockNetCard(name string) MockNetCard {
	return MockNetCard{
		Addr: name,
	}
}

// Empty 实现NetInterface
func (card MockNetCard) Empty() bool {
	return card.In.Empty()
}

// Recv 实现NetInterface
func (card MockNetCard) Recv() (Message, error) {
	return card.In.Dequeue()
}

// Send 实现NetInterface
func (card MockNetCard) Send(m Message) error {
	card.Out.InQueue(m)
	return nil
}

// GetAddr 实现NetInterface
func (card MockNetCard) GetAddr() string {
	return card.Addr
}

// JoinNetWork 连接到MockNetwork
func (card *MockNetCard) JoinNetWork(net *MockNetwork) {
	card.In = &Vec[Message]{}
	card.Out = &Vec[Message]{}
	net.Outs[card.Addr] = card.In
	net.Ins[card.Addr] = card.Out
}

// MockNetwork 模拟的网络组件
type MockNetwork struct {
	Os         OsApi
	NetLatency int32
	Waittings  Vec[Message]
	Ins        map[string]*Vec[Message]
	Outs       map[string]*Vec[Message]
}

// NewMockNetWork 创建新的模拟网络组件
func NewMockNetWork(latency int32) *MockNetwork {
	return &MockNetwork{
		NetLatency: latency,
		Waittings:  Vec[Message]{},
		Ins:        make(map[string]*Vec[Message]),
		Outs:       make(map[string]*Vec[Message]),
	}
}

// Component 实现 NodeComponent
func (n MockNetwork) Component() ComponentName { return CMockNetWork }

// SetOsApi 实现 NodeComponent
func (n *MockNetwork) SetOsApi(osapi OsApi) { n.Os = osapi }

// String 用于Debug
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

// EcsRunCluster 创建ECS ，为ECS添加模拟的网络组件 ， 为集群所有组件提供模拟的系统调用
// 将cluster的的Setups Updates 转换为ECS 统一更新的system.
// 运行集群
func EcsRunCluster(cluster Cluster) {
	simulator := NewEcs()
	newNet := NewMockNetWork(Config.NetLatency)
	getTimeFunc := func() time.Time { return ZEROTIME.Add(time.Duration(simulator.UpdateCount) * time.Microsecond * 100) }
	card := CreateMockNetCard("network1" + ":" + string(CMockNetWork))
	card.JoinNetWork(newNet)
	newNet.SetOsApi(
		&MockOs{
			getTimeFunc,
			card,
		},
	)

	simulator.AddEntities("network1", newNet)
	simulator.AddSystem(SystemName(string(newNet.Component()+"_update")), func(e *ECS) {
		e.ApplyToAllComponent(CMockNetWork, networkTick)
	})

	for _, node := range cluster.Nodes {
		var inited []Component
		for _, c := range node.Components {

			card := CreateMockNetCard(node.Name + ":" + string(c.Component()))
			card.JoinNetWork(newNet)

			c.SetOsApi(
				&MockOs{
					getTimeFunc,
					card,
				},
			)

			inited = append(inited, c)
		}
		simulator.AddEntities(EntityName(node.Name), inited...)
	}

	// 组件 方法装换为 ECS 的system

	for k, f := range cluster.Setups {
		simulator.AddSystem(SystemName(string(k)+"_setup"), covertFuncToSystem(k, f, true))
	}

	for k, f := range cluster.Updates {
		simulator.AddSystem(SystemName(string(k)+"_update"), covertFuncToSystem(k, f, false))
	}

	// 运行 200000 帧 ，每帧间隔0.1ms 模拟 20 秒的集群行为
	frameNum := 200000
	for i := 0; i < frameNum; i++ {
		log.Println("simluating", i, "/", frameNum)
		simulator.Update()

	}

}

func covertFuncToSystem(c ComponentName, f func(interface{}), isSetup bool) func(e *ECS) {
	return func(e *ECS) {
		if isSetup {
			if e.UpdateCount != 0 {
				return
			}
		}

		componetTick := func(ecs *ECS, e EntityName, comp Component) Component {
			switch t := comp.(type) {
			case *TaskGen:
				f(t)
				return t
			case *Scheduler:
				f(t)
				return t
			case *ResourceManager:
				f(t)
				return t

			case *StateStorage:
				f(t)
				return t
			default:
				panic("wrong type componet,if there is new componet type, please add it there")
			}
		}
		e.ApplyToAllComponent(c, componetTick)
	}
}

func networkTick(_ *ECS, _ EntityName, comp Component) Component {
	n := comp.(*MockNetwork)

	for _, in := range n.Ins {
		for !in.Empty() {
			newM, err := in.Dequeue()
			// message body can not be pointer
			common.AssertTypeIsNotPointer(newM.Body)
			if common.IsSameHost(newM.To, newM.From) {
				newM.LeftTime = 0
			} else {
				newM.LeftTime = common.RandIntWithRange(n.NetLatency*10, 0.3)
			}

			if err != nil {
				panic(err)
			}

			n.Waittings.InQueue(newM)
		}

	}
	for i := 0; i < len(n.Waittings); {
		m := n.Waittings[i]
		needDelete := false
		if m.LeftTime == 0 {
			LogInfo(n.Os, ": new message sended", m.From, m.To, m.Content)
			NetEventLog(_getTime_ms(n.Os), "sended", &m)
			out, ok := n.Outs[m.To]
			if !ok {
				panic(m.To + ":net can not reach")
			}
			needDelete = true
			out.InQueue(m)
		} else {
			n.Waittings[i].LeftTime--
		}
		if needDelete {
			n.Waittings.Delete(i)
		} else {
			i++
		}

	}

	return n

}
func _getTime_ms(osapi OsApi) int64 {
	return osapi.GetTime().Sub(ZEROTIME).Milliseconds()
}

// LogInfo 在ECS运行组件的日志处理
func LogInfo(osapi OsApi, ins ...interface{}) {
	s := fmt.Sprint(_getTime_ms(osapi), " ", "Info", " ", osapi.Net().GetAddr(), " ")
	for _, item := range ins {
		s += fmt.Sprint(item, " ")
	}
	fmt.Println(s)
}
