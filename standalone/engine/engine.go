package engine

import (
	"fmt"
	"log"

	"time"

	"simds-standalone/cluster/base"
	"simds-standalone/common"
	"simds-standalone/config"
	"simds-standalone/tracing/rules"
)

// ZEROTIME 模拟开始的现实时间，以此作为模拟器的零点时间
var ZEROTIME time.Time = time.Now()

type Progress uint32

const FullProgress Progress = 1000000

func (p *Progress) toFloat() float32 {
	return float32(*p) / float32(FullProgress)
}

func (p *Progress) Add(percent float32) {
	*p += Progress(percent * float32(FullProgress))
}

func (p *Progress) IsFinished() bool {
	return *p >= FullProgress
}

type ActorHideStatus struct {
	IsBusy     bool
	Progress   Progress
	LastMsg    *base.Message
	Difficulty time.Duration
}

func (actor *EngineActor) UpdateProgress(t time.Time, percent float32) {
	actor.hide.Progress.Add(percent)
	if actor.hide.Progress.IsFinished() {
		rules.CheckRulesThenExec(rules.MsgFinishRules, t, actor.hide.LastMsg)
		actor.hide.Progress = 0
		actor.hide.IsBusy = false
	}
}

func (hide *ActorHideStatus) ToBusy(msg *base.Message, difficulty time.Duration) {
	hide.IsBusy = true
	hide.Progress = 0
	hide.Difficulty = difficulty
	hide.LastMsg = msg
}

type EngineActor struct {
	model base.Actor
	hide  ActorHideStatus
}

type VirtualNode struct {
	engine     *Engine
	updatefunc func([]ActorHideStatus, ActorHideStatus) float32
	actors     map[string]*EngineActor
}

func _defaultUpdateFunc(_ []ActorHideStatus, self ActorHideStatus) float32 {
	delta := time.Second / time.Duration(config.Val.FPS)
	return float32(delta) / float32(self.Difficulty)
}

func NewVirtualNode(engine *Engine, actors ...base.Actor) *VirtualNode {
	var vnode VirtualNode
	vnode.engine = engine
	vnode.updatefunc = _defaultUpdateFunc
	vnode.actors = make(map[string]*EngineActor)
	for _, actor := range actors {
		vnode.actors[actor.GetHostName()] = &EngineActor{actor, ActorHideStatus{}}
	}
	return &vnode
}

func (vnode *VirtualNode) Update() {
	var lastState []ActorHideStatus
	for _, actor := range vnode.actors {
		lastState = append(lastState, actor.hide)
	}

	for _, actor := range vnode.actors {
		if actor.hide.IsBusy {
			actor.UpdateProgress(vnode.engine.GetWorldTime(), vnode.updatefunc(lastState, actor.hide))
		} else if msg, err := vnode.engine.Network.Outs[actor.model.GetHostName()].Dequeue(); err == nil {
			t := time.Now()
			actor.model.Update(msg)   // update the data status of the actor
			costTime := time.Since(t) // record the time cost, the value is treat as the task's "difficulty"
			rules.CheckRulesThenExec(rules.MsgDealRules, vnode.engine.GetWorldTime(), &msg)

			actor.hide.ToBusy(&msg, costTime)
		}
		actor.model.SimulateTasksUpdate() // 模拟任务进度更新。
	}
}

// MockOs 为组件提供模拟的系统调用
type EngineOs struct {
	addr   string
	engine *Engine
}

// GetTime 提供模拟时间
func (o *EngineOs) GetTime() time.Time {
	return o.engine.GetWorldTime()
}

func (o *EngineOs) Run(f func()) {
}

func (o *EngineOs) Send(m base.Message) error {
	m.Id = fmt.Sprint(o.GetTime().UnixMicro()) + "_" + m.From + "_" + m.To
	o.engine.Network.Ins[o.addr].InQueue(m)

	rules.CheckRulesThenExec(rules.SendRules, o.GetTime(), &m)

	return nil
}

//func (o *EngineOs) LogInfo(out string, items ...string) {
//	timestr := o.GetTime().Format(time.RFC3339Nano)
//	s := ""
//	if out == "stdout" {
//		s += fmt.Sprint(timestr)
//		for _, item := range items {
//			s += ","
//			s += fmt.Sprint(item)
//		}
//		fmt.Println(s)
//	} else {
//		line := append([]string{timestr}, items...)
//		err := common.AppendLineCsvFile(path.Join(config.Val.OutputDir, out), line)
//		if err != nil {
//			panic(err)
//		}
//	}
//}

// MockNetwork 模拟的网络组件
type VirtualNetwork struct {
	Os         base.OsApi
	NetLatency int32
	Waittings  base.Vec[base.Message]
	Ins        map[string]*base.Vec[base.Message]
	Outs       map[string]*base.Vec[base.Message]
}

func newVirtualNetwork() VirtualNetwork {
	return VirtualNetwork{
		NetLatency: config.Val.NetLatency,
		Waittings:  base.Vec[base.Message]{},
		Ins:        make(map[string]*base.Vec[base.Message]),
		Outs:       make(map[string]*base.Vec[base.Message]),
	}
}

type Engine struct {
	UpdateCount uint64
	UpdateGap   time.Duration // 每次更新推进的时间
	Nodes       []VirtualNode
	Network     VirtualNetwork
}

// 集群已经运行的时间
func (engine *Engine) UpTime() time.Duration {
	return time.Duration(engine.UpdateCount) * engine.UpdateGap
}

func (engine *Engine) GetWorldTime() time.Time {
	return ZEROTIME.Add(engine.UpTime())
}

// 并行地对所有节点更新一次
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
				engine.Nodes[j].Update()
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

// 对单个节点运行节点更新函数
//func (engine *Engine) _updateNode(node *base.Node) {
//	for _, actor := range node.Actors {
//		addr := actor.GetHostName()
//		nextUpdateTime := actor.GetNextUpdateTime()
//
//		if engine.GetWorldTime().Sub(nextUpdateTime) >= 0 {
//			if msg, err := engine.Network.Outs[addr].Dequeue(); err == nil {
//				t := time.Now()
//				actor.Update(msg) // 事件循环处理。
//				costTime := time.Since(t)
//				actor.SetNextUpdateTime(engine.GetWorldTime().Add(costTime)) // 设置下一次更新的时间
//				//engine.Nodes[j].SetNextUpdateTime(engine.GetWorldTime().Add(time.Millisecond)) // 设置下一次更新的时间
//
//			}
//		}
//		actor.SimulateTasksUpdate() // 模拟任务进度更新。
//	}
//}

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
			out.InQueue(m)
			rules.CheckRulesThenExec(rules.RecvRules, engine.GetWorldTime(), &m)
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

func InitEngine(cluster base.Cluster) *Engine {
	rules.InitTracing()

	var e Engine
	for _, node := range cluster.Nodes {
		e.Nodes = append(e.Nodes, *NewVirtualNode(&e, node.Actors...))
	}

	e.Network = newVirtualNetwork()
	e.Network.Os = &EngineOs{addr: "network", engine: &e}
	e.UpdateGap = time.Second / time.Duration(config.Val.FPS)
	for _, node := range e.Nodes {
		for _, actor := range node.actors {
			e.Network.Ins[actor.model.GetHostName()] = &base.Vec[base.Message]{}
			e.Network.Outs[actor.model.GetHostName()] = &base.Vec[base.Message]{}
		}
	}

	for i := range e.Nodes {
		for _, actor := range e.Nodes[i].actors {
			os := EngineOs{}
			os.addr = actor.model.GetHostName()
			os.engine = &e

			os.Send(base.Message{
				From:    os.addr,
				To:      os.addr,
				Content: "SignalBoot",
				Body:    base.Signal("SignalBoot"),
			})
			actor.model.SetOsApi(&os)
		}
	}
	return &e
}

func (engine *Engine) Run() {
	testDuration := (time.Duration(config.Val.SimulateDuration) * (time.Millisecond)).Seconds()

	frameNum := (testDuration + 15.0) * float64(config.Val.FPS)

	step := uint64(config.Val.FPS)
	for engine.UpdateCount < uint64(frameNum) {
		log.Println("Simulation Progress", time.Duration(engine.UpdateCount*uint64(time.Second)/uint64(config.Val.FPS)).Milliseconds(), "ms", "/", config.Val.SimulateDuration, "ms")
		start := time.Now()
		engine.UpdateNtimes(step)
		log.Println("render", step, "frames", "spend:", time.Since(start))
	}

}
