package core

//ecs_runner.go 使用 通用的ecs.go定义的通用ECS来模拟XXXCluster

import (
	"fmt"
	"io"
	"log"
	"simds-standalone/common"
	"simds-standalone/config"
	"strings"
	"time"

	"github.com/chzyer/readline"
	lua "github.com/yuin/gopher-lua"
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

// SetOsApi 实现 NodeComponent
func (n *MockNetwork) Debug() { log.Println(n.Waittings) }

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

func newSimulator(cluster Cluster) *ECS {
	simulator := NewEcs()
	newNet := NewMockNetWork(config.Val.NetLatency)
	getTimeFunc := func() time.Time { return ZEROTIME.Add(time.Duration(simulator.UpdateCount*1000000/uint64(config.Val.FPS))*time.Microsecond)}
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
	return simulator

}

func covertFuncToSystem(c ComponentName, f func(Component), isSetup bool) func(e *ECS) {
	return func(e *ECS) {
		if isSetup {
			if e.UpdateCount != 0 {
				return
			}
		}

		componetTick := func(ecs *ECS, e EntityName, comp Component) Component {
			f(comp)
			return comp
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
				newM.LeftTime = time.Duration(common.RandIntWithRange(n.NetLatency*1000, 0.3)) * time.Microsecond
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
		if m.LeftTime < 0 {
			LogInfo(n.Os, ": new message sended", m.From, m.To, m.Content)
			NetEventLog(_getTime_ms(n.Os), "sended", &m)
			out, ok := n.Outs[m.To]
			if !ok {
				panic(m.To + ":net can not reach")
			}
			needDelete = true
			out.InQueue(m)
		} else {
			n.Waittings[i].LeftTime -= (time.Second / time.Duration(config.Val.FPS))
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

// EcsRunCluster 创建ECS ，为ECS添加模拟的网络组件 ， 为集群所有组件提供模拟的系统调用
// 将cluster的的Setups Updates 转换为ECS 统一更新的system.
// 运行集群
func EcsRunCluster(cluster Cluster) {
	simulator := newSimulator(cluster)

	frameNum := uint64(config.Val.FPS * config.Val.SimulateDuration / 1000)

	step := uint64(config.Val.FPS)
	for simulator.UpdateCount < uint64(frameNum) {
		log.Println("Simulation Progress", simulator.UpdateCount*1000/uint64(config.Val.FPS), "ms", "/", config.Val.SimulateDuration, "ms")
		simulator.UpdateNtimes(step)
	}
}

func simdsLua(simulator *ECS) *lua.LState {
	l := lua.NewState()
	step := func(L *lua.LState) int {
		lv := L.ToInt(1) /* get argument */
		simulator.UpdateNtimes(uint64(lv))
		return 1 /* number of results */
	}

	show := func(L *lua.LState) int {
		arg1 := L.ToString(1) /* get argument */
		switch {
		case arg1 == "cluster":
			log.Println("cluster is type:", config.Val.Cluster)
		case arg1 == "config":
			fmt.Println(config.LogString())
		case arg1 == "hosts":
			keys := make([]EntityName, 0, len(simulator.Entities))
			for en := range simulator.Entities {
				keys = append(keys, en)
			}
			fmt.Println(keys)
		default:
			simulator.ShowEntities(EntityName(arg1))
		}

		return 1 /* number of results */
	}

	time := func(L *lua.LState) int {
		log.Println("Simulator Time: ", simulator.UpdateCount*1000*uint64(config.Val.FPS), "ms")
		return 1
	}

	l.SetGlobal("step", l.NewFunction(step))
	l.SetGlobal("show", l.NewFunction(show))
	l.SetGlobal("time", l.NewFunction(time))
	return l
}

// Debug模式下进行运行集群
func EcsRunClusterDebug(cluster Cluster) {
	simulator := newSimulator(cluster)

	luaState := simdsLua(simulator)
	defer luaState.Close()

	l, err := readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf(">>> "),
		HistoryFile:     "/tmp/readline.tmp",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold: true,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()
	l.CaptureExitSignal()

	log.SetOutput(l.Stderr())
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if err := luaState.DoString(line); err != nil {
			log.Println(err)
		}
	}
}
