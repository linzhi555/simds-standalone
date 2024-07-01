package engine

import (
	"fmt"
	"io"
	"log"
	"path"
	"simds-standalone/common"
	"simds-standalone/config"
	"simds-standalone/core"
	"strings"
	"time"

	"github.com/chzyer/readline"
	lua "github.com/yuin/gopher-lua"
)

const NETWORK_EVENT_LOG_NAME = "network_event.log"

// ZEROTIME 模拟开始的现实时间，以此作为模拟器的零点时间
var ZEROTIME time.Time = time.Now()

// MockOs 为组件提供模拟的系统调用
type EngineOs struct {
	host   string
	engine *Engine
}

// GetTime 提供模拟时间
func (o *EngineOs) GetTime() time.Time {
	return o.engine.GetWorldTime()
}

func (o *EngineOs) Run(f func()) {

}

// func (o *EngineOs) HasMessage() bool {
// 	return !o.engine.Network.Outs[o.host].Empty()
// }

func (o *EngineOs) Recv() (core.Message, error) {
	return o.engine.Network.Outs[o.host].Dequeue()
}

func (o *EngineOs) Send(m core.Message) error {
	o.engine.Network.Ins[o.host].InQueue(m)
	return nil
}

func (o *EngineOs) LogInfo(out string, items ...string) {
	timestr := o.GetTime().Format(time.RFC3339Nano)
	s := ""
	if out == "stdout" {
		s += fmt.Sprint(timestr)
		for _, item := range items {
			s += ","
			s += fmt.Sprint(item)
		}
		fmt.Println(s)
	} else {
		line := append([]string{timestr}, items...)
		err := common.AppendLineCsvFile(path.Join(config.Val.OutputDir, out), line)
		if err != nil {
			panic(err)
		}
	}
}

// MockNetwork 模拟的网络组件
type VirtualNetwork struct {
	Os         core.OsApi
	NetLatency int32
	Waittings  core.Vec[core.Message]
	Ins        map[string]*core.Vec[core.Message]
	Outs       map[string]*core.Vec[core.Message]
}

func newVirtualNetwork() VirtualNetwork {
	return VirtualNetwork{
		NetLatency: config.Val.NetLatency,
		Waittings:  core.Vec[core.Message]{},
		Ins:        make(map[string]*core.Vec[core.Message]),
		Outs:       make(map[string]*core.Vec[core.Message]),
	}
}

type Engine struct {
	UpdateCount uint64
	UpdateGap   time.Duration // 每次更新推进的时间
	Nodes       []core.Node
	Network     VirtualNetwork
}

func (engine *Engine) GetWorldTime() time.Time {
	return ZEROTIME.Add(time.Duration(engine.UpdateCount) * engine.UpdateGap)
}

// / 对所有节点更新一次
func (engine *Engine) updateNodes() {
	allNum := len(engine.Nodes)

	RenderThreadNum := int(config.Val.GoProcs)
	if allNum < int(config.Val.GoProcs) {
		RenderThreadNum = allNum
	}

	finishChan := make(chan bool, RenderThreadNum)

	partNum := allNum / RenderThreadNum
	start := 0
	end := start + partNum
	for i := 0; i < RenderThreadNum; i++ {
		go func(s, e int) {
			for j := s; j < e; j++ {
				hostname := engine.Nodes[j].GetHostName()
				nextUpdateTime := engine.Nodes[j].GetNextUpdateTime()

				if engine.GetWorldTime().Sub(nextUpdateTime) >= 0 {
					if msg, err := engine.Network.Outs[hostname].Dequeue(); err == nil {
						t := time.Now()
						engine.Nodes[j].Update(msg) // 事件循环处理。
						costTime := time.Since(t)
						engine.Nodes[j].SetNextUpdateTime(engine.GetWorldTime().Add(costTime)) // 设置下一次更新的时间
						//engine.Nodes[j].SetNextUpdateTime(engine.GetWorldTime().Add(time.Millisecond)) // 设置下一次更新的时间

					}
				}
				engine.Nodes[j].SimulateTasksUpdate() // 模拟任务进度更新。
			}
			finishChan <- true
		}(start, end)
		start = end
		if i == RenderThreadNum-2 {
			end = allNum
		} else {
			end = start + partNum
		}
	}

	for i := 0; i < RenderThreadNum; i++ {
		<-finishChan
	}

}

// 对集群引擎的虚拟网络进行更新
func (engine *Engine) updateNetwork() {
	network := &engine.Network
	for _, in := range network.Ins {
		for !in.Empty() {
			newM, err := in.Pop()
			// message body can not be pointer
			common.AssertTypeIsNotPointer(newM.Body)
			if newM.To == newM.From {
				newM.LeftTime = 0
			} else {
				newM.LeftTime = time.Duration(common.RandIntWithRange(network.NetLatency*1000, 0.3)) * time.Microsecond
			}
			if err != nil {
				panic(err)
			}

			network.Waittings.InQueue(newM)
		}
	}
	for i := 0; i < len(network.Waittings); {
		m := network.Waittings[i]
		needDelete := false
		if m.LeftTime < 0 {
			out, ok := network.Outs[m.To]
			if !ok {
				panic(fmt.Sprint(m) + ":net can not reach")
			}
			needDelete = true
			network.Os.LogInfo(NETWORK_EVENT_LOG_NAME, m.Content, m.From, m.To)
			out.InQueue(m)
		} else {
			network.Waittings[i].LeftTime -= (time.Second / time.Duration(config.Val.FPS))
		}
		if needDelete {
			network.Waittings.Delete(i)
		} else {
			i++
		}
	}

}

// 对集群更新一次状态，推进一个单位时间
func (engine *Engine) Update() {
	engine.updateNodes()
	engine.updateNetwork()
	engine.UpdateCount++
}

func (engine *Engine) UpdateNtimes(n uint64) {
	for i := uint64(0); i < n; i++ {
		engine.Update()
	}
}

func InitEngine(cluster core.Cluster) *Engine {
	common.AppendLineCsvFile(path.Join(config.Val.OutputDir, NETWORK_EVENT_LOG_NAME), []string{"time", "type", "from", "to"})
	common.AppendLineCsvFile(path.Join(config.Val.OutputDir, core.TASKS_EVENT_LOG_NAME), []string{"time", "taskid", "type", "nodeip", "cpu", "ram"})

	var e Engine
	e.Nodes = cluster.Nodes
	e.Network = newVirtualNetwork()
	e.Network.Os = &EngineOs{host: "network", engine: &e}
	e.UpdateGap = time.Second / time.Duration(config.Val.FPS)
	for _, node := range e.Nodes {
		e.Network.Ins[node.GetHostName()] = &core.Vec[core.Message]{}
		e.Network.Outs[node.GetHostName()] = &core.Vec[core.Message]{}
	}

	for i := range e.Nodes {
		os := EngineOs{}
		os.host = e.Nodes[i].GetHostName()
		os.engine = &e

		os.Send(core.Message{
			From:    os.host,
			To:      os.host,
			Content: "SignalBoot",
			Body:    core.Signal("SignalBoot"),
		})

		e.Nodes[i].SetNextUpdateTime(e.GetWorldTime())
		e.Nodes[i].SetOsApi(&os)
	}

	return &e
}

func (engine *Engine) Run() {
	frameNum := (time.Duration(config.Val.SimulateDuration) * (time.Millisecond)).Seconds() * float64(config.Val.FPS)

	step := uint64(config.Val.FPS)
	for engine.UpdateCount < uint64(frameNum) {
		log.Println("Simulation Progress", time.Duration(engine.UpdateCount*uint64(time.Second)/uint64(config.Val.FPS)).Milliseconds(), "ms", "/", config.Val.SimulateDuration, "ms")
		start := time.Now()
		engine.UpdateNtimes(step)
		log.Println("render", step, "frames", "spend:", time.Since(start))
	}

}

// Console模式下进行运行集群
func (engine *Engine) RunInConsole() {

	luaState := simdsLua(engine)
	defer luaState.Close()

	l, err := readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf("\033[32m>>> \033[0m"),
		HistoryFile:     "/tmp/readline.tmp",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold: true,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()
	//l.CaptureExitSignal()

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

func simdsLua(simulator *Engine) *lua.LState {
	l := lua.NewState()
	step := func(L *lua.LState) int {
		lv := L.ToInt(1) /* get argument */
		simulator.UpdateNtimes(uint64(lv))
		return 1 /* number of results */
	}

	to := func(L *lua.LState) int {
		lv := L.ToInt(1) /* get argument */
		simulator.UpdateNtimes(uint64(lv) - simulator.UpdateCount)
		return 1 /* number of results */
	}


	time := func(L *lua.LState) int {
		fmt.Printf("Simulator Time: %f s, UpdateFrames: %d, FPS: %d \n", float32(simulator.UpdateCount)/float32(config.Val.FPS), simulator.UpdateCount, config.Val.FPS)
		return 1
	}

	l.SetGlobal("step", l.NewFunction(step))
	l.SetGlobal("to", l.NewFunction(to))
	l.SetGlobal("time", l.NewFunction(time))
	return l
}
