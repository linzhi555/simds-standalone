package engine

import (
	"fmt"
	"log"
	"path"
	"simds-standalone/common"
	"simds-standalone/config"
	"simds-standalone/core"
	"time"
)

const NETWORK_EVENT_LOG_NAME = "network_event.log"

// ZEROTIME 模拟开始的现实时间，以此作为模拟器的零点时间
var ZEROTIME time.Time = time.Now()

// MockOs 为组件提供模拟的系统调用
type EngineOs struct {
	host   string
	node   core.Node
	engine *Engine
	tasks  *core.Vec[core.TaskInfo]
}

// GetTime 提供模拟时间
func (o *EngineOs) GetTime() time.Time {
	return ZEROTIME.Add(time.Duration(o.engine.UpdateCount*1000000/uint64(config.Val.FPS)) * time.Microsecond)
}

func (o *EngineOs) Run(f func()) {

}

func (o *EngineOs) HasMessage() bool {
	return !o.engine.Network.Outs[o.host].Empty()
}

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
	Nodes       []core.Node
	Network     VirtualNetwork
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
				t := time.Now()

				engine.Nodes[j].Update() // 事件循环处理。
				if engine.Nodes[j].GetHostName() == "master0" {
					log.Println(time.Since(t))
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

	for _, node := range e.Nodes {
		e.Network.Ins[node.GetHostName()] = &core.Vec[core.Message]{}
		e.Network.Outs[node.GetHostName()] = &core.Vec[core.Message]{}
	}

	for i := range e.Nodes {
		os := EngineOs{}
		os.host = e.Nodes[i].GetHostName()
		os.engine = &e
		e.Nodes[i].SetOsApi(&os)
	}
	return &e
}

func (engine *Engine) Run() {
	frameNum := uint64(config.Val.FPS * config.Val.SimulateDuration / 1000)
	step := uint64(config.Val.FPS)
	for engine.UpdateCount < uint64(frameNum) {
		log.Println("Simulation Progress", engine.UpdateCount*1000/uint64(config.Val.FPS), "ms", "/", config.Val.SimulateDuration, "ms")
		start := time.Now()
		engine.UpdateNtimes(step)
		log.Println("render", step, "frames", "spend:", time.Since(start))
	}

}
