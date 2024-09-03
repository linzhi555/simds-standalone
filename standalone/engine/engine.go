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
		actor.hide.Difficulty = 0
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
	m.Id = common.GenerateUID()
	o.engine.Network.Ins[o.addr].InQueueBack(m)

	rules.CheckRulesThenExec(rules.SendRules, o.GetTime(), &m)

	return nil
}

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

// 对集群引擎的虚拟网络进行更新
func (engine *Engine) updateNetwork() {
	network := &engine.Network

	// 在上次更新中产生的消息要被运输到集中存储区域处理
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

			network.Waittings.InQueueBack(newM)
		}
	}

	// 集中处理message vector比map[ActorId]Messge速度更快
	for i := 0; i < len(network.Waittings); {
		m := network.Waittings[i]
		needDelete := false
		if m.LeftTime < 0 {
			out, ok := network.Outs[m.To]
			if !ok {
				panic(fmt.Sprint(m) + ":net can not reach")
			}
			needDelete = true
			out.InQueueBack(m)
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

		start := time.Now()
		engine.UpdateNtimes(step)

		log.Printf(
			"%.4f%% progress:( %d / %d) current speed:%v / %d frame  FPS:%.1f\n",
			float32(engine.UpdateCount)/float32(frameNum)*100.0,
			engine.UpdateCount,
			int64(frameNum),
			time.Since(start), step,
			float64(step)/time.Since(start).Seconds(),
		)

		if engine.UpdateCount == 25*step {
			common.MemProf()
		}

	}
}

type ActorDebugInfo struct {
	Name       string `json:"name"`
	Node       string `json:"node"`
	IsBusy     string `json:"isBusy"`
	Progress   string `json:"progress"`
	Msg        string `json:"msg"`
	Difficulty string `json:"difficulty"`
	Data       string `json:"data"`
}

func (engine *Engine) DebugNodes() []ActorDebugInfo {
	var res []ActorDebugInfo
	for _, node := range engine.Nodes {
		for _, actor := range node.actors {
			res = append(res, ActorDebugInfo{
				Name:     actor.model.GetHostName(),
				Node:     actor.model.GetHostName(),
				IsBusy:   fmt.Sprint(actor.hide.IsBusy),
				Progress: fmt.Sprint(actor.hide.Progress.toFloat()),

				Msg: func() string {
					if actor.hide.LastMsg == nil {
						return "null"
					}
					return actor.hide.LastMsg.Content
				}(),

				Difficulty: fmt.Sprint(actor.hide.Difficulty),
			})
		}
	}
	return res
}

type MessageDebugInfo struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Head     string `json:"head"`
	Body     string `json:"body"`
	LeftTime string `json:"leftTime"`
}

type NetDebugInfo struct {
	Waitings []MessageDebugInfo `json:"waittings"`
	Sended   []MessageDebugInfo `json:"sended"`
}

func (engine *Engine) DebugNet() NetDebugInfo {
	var res NetDebugInfo

	addMsg := func(target *[]MessageDebugInfo, msg *base.Message) {
		*target = append(*target, MessageDebugInfo{
			From:     msg.From,
			To:       msg.To,
			Head:     msg.Content,
			Body:     "",
			LeftTime: fmt.Sprint(msg.LeftTime),
		})
	}

	for i := range engine.Network.Waittings {
		addMsg(&res.Waitings, &engine.Network.Waittings[i])
	}

	for _, ins := range engine.Network.Ins {
		for i := range *ins {
			addMsg(&res.Sended, &(*ins)[i])
		}
	}

	log.Println(len(res.Sended), len(res.Waitings))

	return res
}
