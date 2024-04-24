package engine

// //ecs_runner.go 使用 通用的ecs.go定义的通用ECS来模拟XXXCluster

// import (
// 	"fmt"
// 	"io"
// 	"log"
// 	"simds-standalone/common"
// 	"simds-standalone/config"
// 	"simds-standalone/core"
// 	"strings"
// 	"time"

// 	"github.com/chzyer/readline"
// 	lua "github.com/yuin/gopher-lua"
// )

// // JoinNetWork 连接到MockNetwork
// func (card *MockNetCard) JoinNetWork(net *MockNetwork) {
// 	card.In = &core.Vec[core.Message]{}
// 	card.Out = &core.Vec[core.Message]{}
// 	net.Outs[card.Addr] = card.In
// 	net.Ins[card.Addr] = card.Out
// }

// // NewMockNetWork 创建新的模拟网络组件
// func NewMockNetWork(latency int32) *MockNetwork {
// 	return &MockNetwork{
// 		NetLatency: latency,
// 		Waittings:  core.Vec[core.Message]{},
// 		Ins:        make(map[string]*core.Vec[core.Message]),
// 		Outs:       make(map[string]*core.Vec[core.Message]),
// 	}
// }

// // Component 实现 NodeComponent
// func (n MockNetwork) Component() core.ComponentName { return core.CMockNetWork }

// // SetOsApi 实现 NodeComponent
// func (n *MockNetwork) SetOsApi(osapi core.OsApi) { n.Os = osapi }

// // SetOsApi 实现 NodeComponent
// func (n *MockNetwork) Debug() {
// 	fmt.Println("messages in networks:")
// 	for _, m := range n.Waittings {
// 		fmt.Printf("%+v\n", m)
// 	}
// }

// func newSimulator(cluster core.Cluster) *ECS {
// 	simulator := NewEcs()
// 	newNet := NewMockNetWork(config.Val.NetLatency)
// 	getTimeFunc := func() time.Time {
// 		return ZEROTIME.Add(time.Duration(simulator.UpdateCount*1000000/uint64(config.Val.FPS)) * time.Microsecond)
// 	}
// 	card := CreateMockNetCard("network1" + ":" + string(core.CMockNetWork))
// 	card.JoinNetWork(newNet)
// 	newNet.SetOsApi(
// 		&MockOs{
// 			getTimeFunc,
// 			card,
// 		},
// 	)

// 	simulator.AddEntities("network1", newNet)
// 	simulator.AddSystem(SystemName(string(newNet.Component()+"_update")), func(e *ECS) {
// 		e.ApplyToAllComponent(core.CMockNetWork, networkTick)
// 	})

// 	for _, node := range cluster.Nodes {
// 		var inited []core.Component
// 		for _, c := range node.Components {

// 			card := CreateMockNetCard(node.Name + ":" + string(c.Component()))
// 			card.JoinNetWork(newNet)

// 			c.SetOsApi(
// 				&MockOs{
// 					getTimeFunc,
// 					card,
// 				},
// 			)

// 			inited = append(inited, c)
// 		}
// 		simulator.AddEntities(EntityName(node.Name), inited...)
// 	}

// 	// 初始化所有的节点
// 	for c, f := range cluster.Setups {
// 		componetTick := func(ecs *ECS, e EntityName, comp core.Component) core.Component {
// 			f(comp)
// 			return comp
// 		}
// 		simulator.ApplyToAllComponent(c, componetTick)
// 	}

// 	NodesUpdatesFunc := func(e *ECS) {
// 		finishChan := make(chan bool, len(cluster.Updates))
// 		for componetname, updatefunc := range cluster.Updates {
// 			go func(c core.ComponentName, f func(core.Component)) {
// 				componetTick := func(ecs *ECS, e EntityName, comp core.Component) core.Component {
// 					f(comp)
// 					return comp
// 				}
// 				e.ApplyToAllComponent(c, componetTick)
// 				finishChan <- true
// 			}(componetname, updatefunc)
// 		}
// 		for i := 0; i < len(cluster.Updates); i++ {
// 			<-finishChan
// 		}
// 	}
// 	simulator.AddSystem(SystemName("nodes_update"), NodesUpdatesFunc)
// 	return simulator

// }

// func networkTick(_ *ECS, _ EntityName, comp core.Component) core.Component {
// 	n := comp.(*MockNetwork)
// 	var t1, t2, t3 time.Time
// 	t1 = time.Now()
// 	for _, in := range n.Ins {
// 		for !in.Empty() {
// 			newM, err := in.Pop()
// 			// message body can not be pointer
// 			common.AssertTypeIsNotPointer(newM.Body)
// 			if common.IsSameHost(newM.To, newM.From) {
// 				newM.LeftTime = 0
// 			} else {
// 				newM.LeftTime = time.Duration(common.RandIntWithRange(n.NetLatency*1000, 0.3)) * time.Microsecond
// 			}

// 			if err != nil {
// 				panic(err)
// 			}

// 			n.Waittings.InQueue(newM)
// 		}
// 	}
// 	t2 = time.Now()
// 	for i := 0; i < len(n.Waittings); {
// 		m := n.Waittings[i]
// 		needDelete := false
// 		if m.LeftTime < 0 {
// 			LogInfo(n.Os, ": new message sended", m.From, m.To, m.Content)
// 			core.NetEventLog(_getTime_ms(n.Os), "sended", &m)
// 			out, ok := n.Outs[m.To]
// 			if !ok {
// 				panic(m.To + ":net can not reach")
// 			}
// 			needDelete = true
// 			out.InQueue(m)
// 		} else {
// 			n.Waittings[i].LeftTime -= (time.Second / time.Duration(config.Val.FPS))
// 		}
// 		if needDelete {
// 			n.Waittings.Delete(i)
// 		} else {
// 			i++
// 		}
// 	}
// 	t3 = time.Now()
// 	debug := false
// 	if debug {

// 		log.Println(t3.Sub(t2), t2.Sub(t1))
// 	}
// 	return n

// }
// func _getTime_ms(osapi core.OsApi) int64 {
// 	return osapi.GetTime().Sub(ZEROTIME).Milliseconds()
// }

// // LogInfo 在ECS运行组件的日志处理
// func LogInfo(osapi core.OsApi, ins ...interface{}) {
// 	s := fmt.Sprint(_getTime_ms(osapi), " ", "Info", " ", osapi.Net().GetAddr(), " ")
// 	for _, item := range ins {
// 		s += fmt.Sprint(item, " ")
// 	}
// 	fmt.Println(s)
// }

// // EcsRunCluster 创建ECS ，为ECS添加模拟的网络组件 ， 为集群所有组件提供模拟的系统调用
// // 将cluster的的Setups Updates 转换为ECS 统一更新的system.
// // 运行集群
// func EcsRunCluster(cluster core.Cluster) {
// 	simulator := newSimulator(cluster)

// 	frameNum := uint64(config.Val.FPS * config.Val.SimulateDuration / 1000)

// 	step := uint64(config.Val.FPS)
// 	for simulator.UpdateCount < uint64(frameNum) {
// 		log.Println("Simulation Progress", simulator.UpdateCount*1000/uint64(config.Val.FPS), "ms", "/", config.Val.SimulateDuration, "ms")
// 		start := time.Now()
// 		simulator.UpdateNtimes(step)
// 		log.Println("render", step, "frames", "spend:", time.Since(start))
// 	}
// }

// func simdsLua(simulator *ECS) *lua.LState {
// 	l := lua.NewState()
// 	step := func(L *lua.LState) int {
// 		lv := L.ToInt(1) /* get argument */
// 		simulator.UpdateNtimes(uint64(lv))
// 		return 1 /* number of results */
// 	}

// 	to := func(L *lua.LState) int {
// 		lv := L.ToInt(1) /* get argument */
// 		simulator.UpdateNtimes(uint64(lv) - simulator.UpdateCount)
// 		return 1 /* number of results */
// 	}

// 	show := func(L *lua.LState) int {
// 		arg1 := L.ToString(1) /* get argument */
// 		switch {
// 		case arg1 == "cluster":
// 			log.Println("cluster is type:", config.Val.Cluster)
// 		case arg1 == "config":
// 			fmt.Println(config.LogString())
// 		case arg1 == "hosts":
// 			keys := make([]EntityName, 0, len(simulator.Entities))
// 			for en := range simulator.Entities {
// 				keys = append(keys, en)
// 			}
// 			fmt.Println(keys)
// 		default:
// 			simulator.ShowEntities(EntityName(arg1))
// 		}

// 		return 1 /* number of results */
// 	}

// 	time := func(L *lua.LState) int {
// 		fmt.Printf("Simulator Time: %f s, UpdateFrames: %d, FPS: %d \n", float32(simulator.UpdateCount)/float32(config.Val.FPS), simulator.UpdateCount, config.Val.FPS)
// 		return 1
// 	}

// 	l.SetGlobal("step", l.NewFunction(step))
// 	l.SetGlobal("show", l.NewFunction(show))
// 	l.SetGlobal("to", l.NewFunction(to))
// 	l.SetGlobal("time", l.NewFunction(time))
// 	return l
// }

// // Debug模式下进行运行集群
// func EcsRunClusterDebug(cluster core.Cluster) {
// 	simulator := newSimulator(cluster)

// 	luaState := simdsLua(simulator)
// 	defer luaState.Close()

// 	l, err := readline.NewEx(&readline.Config{
// 		Prompt:          fmt.Sprintf("\033[32m>>> \033[0m"),
// 		HistoryFile:     "/tmp/readline.tmp",
// 		InterruptPrompt: "^C",
// 		EOFPrompt:       "exit",

// 		HistorySearchFold: true,
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer l.Close()
// 	l.CaptureExitSignal()

// 	log.SetOutput(l.Stderr())
// 	for {
// 		line, err := l.Readline()
// 		if err == readline.ErrInterrupt {
// 			if len(line) == 0 {
// 				break
// 			} else {
// 				continue
// 			}
// 		} else if err == io.EOF {
// 			break
// 		}

// 		line = strings.TrimSpace(line)
// 		if err := luaState.DoString(line); err != nil {
// 			log.Println(err)
// 		}
// 	}
// }
