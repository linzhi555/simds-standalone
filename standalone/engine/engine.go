package engine

import (
	"context"
	"log"
	"time"

	"simds-standalone/cluster/base"
	"simds-standalone/cluster/lib"
	"simds-standalone/common"
	"simds-standalone/config"
	"simds-standalone/tracing/rules"
)

type renderThread struct {
	ctx  context.Context
	zone []VirtualNode

	startCh  chan struct{}
	finishCh chan struct{}
}

func (t *renderThread) goWorking() {
	t.startCh = make(chan struct{}, 1)
	t.finishCh = make(chan struct{}, 1)

	go func() {
		for {
			select {
			case <-t.ctx.Done():
				log.Println("renderThread finished")
				return
			case <-t.startCh:
				for i := range t.zone {
					t.zone[i].Update()
				}
				t.finishCh <- struct{}{}
			}
		}
	}()
}

type Engine struct {
	// 渲染线程控制
	ctx      context.Context
	stopFunc func()

	// 核心状态
	UpdateCount uint64
	renders     []renderThread
	Network     VirtualNetwork

	// 模拟性能优化记录，与逻辑无关
	NodeUpdateCost time.Duration
	NetUpdateCost  time.Duration
}

func InitEngine(cluster base.Cluster) *Engine {
	rules.InitTracing()

	var e Engine
	e.ctx, e.stopFunc = context.WithCancel(context.Background())
	// init network
	e.Network = newVirtualNetwork()

	// init render threads
	for i := int32(0); i < config.Val.GoProcs; i++ {
		var nodes []VirtualNode

		for j := i; j < int32(len(cluster.Nodes)); j += config.Val.GoProcs {
			node := NewVirtualNode(&e, cluster.Nodes[j].Actors...)
			nodes = append(nodes, *node)
		}

		e.renders = append(e.renders, renderThread{
			ctx:  e.ctx,
			zone: nodes,
		})
	}

	// connect the all the connect to the virtual network
	for _, node := range e.nodes() {
		for i := range node.actors {
			actor := &node.actors[i]

			actor.os = EngineOs{
				addr:   actor.model.GetAddress(),
				engine: &e,
				node:   node,
				In:     &common.Vec[base.Message]{},
				Out:    &common.Vec[base.Message]{},
			}

			actor.model.SetOsApi(&actor.os)

			//net work input is actor's output , output is actor's input
			e.Network.Ins[actor.model.GetAddress()] = actor.os.Out
			e.Network.Outs[actor.model.GetAddress()] = actor.os.In

			actor.os.Send(base.Message{
				From: actor.os.addr,
				To:   actor.os.addr,
				Head: "SignalBoot",
				Body: lib.Signal("SignalBoot"),
			})
		}
	}

	// 开始渲染线程，这些渲染线程会在后台等待，通过信号进行工作。
	for i := range e.renders {
		e.renders[i].goWorking()
	}

	return &e
}

// 让引擎自动模拟运行
func (engine *Engine) Run() {
	testDuration := (time.Duration(config.Val.SimulateDuration) * (time.Millisecond)).Seconds()

	frameNum := (testDuration + 15.0) * float64(config.Val.FPS)

	step := uint64(config.Val.FPS)

	for engine.UpdateCount < uint64(frameNum) {

		start := time.Now()
		engine.UpdateNtimes(step)

		log.Printf(
			"%.4f%% progress:( %d / %d) current speed:%v / %d frame  FPS:%.1f NodeCost:%v NetCost:%v \n",
			float32(engine.UpdateCount)/float32(frameNum)*100.0,
			engine.UpdateCount,
			int64(frameNum),
			time.Since(start), step,
			float64(step)/time.Since(start).Seconds(),
			engine.NodeUpdateCost,
			engine.NetUpdateCost,
		)

		if engine.UpdateCount == 25*step {
			common.MemProf()
		}
	}
}

// 对集群更新一次状态，推进一个单位时间
func (engine *Engine) Update() {

	start := time.Now()
	engine.updateNodes()
	engine.NodeUpdateCost += time.Since(start)

	start = time.Now()
	engine.Network.updateNetwork(engine.GetWorldTime())
	engine.NetUpdateCost += time.Since(start)

	engine.UpdateCount++
}

func (engine *Engine) UpdateNtimes(n uint64) {
	for i := uint64(0); i < n; i++ {
		engine.Update()
	}
}

// 引擎的状态迭代算法

// 给所有渲染线程放送信号，然后等待完成
func (engine *Engine) updateNodes() {
	for i := range engine.renders {
		engine.renders[i].startCh <- struct{}{}
	}

	for i := range engine.renders {
		<-engine.renders[i].finishCh
	}
}

func (engine *Engine) nodes() []*VirtualNode {
	var res []*VirtualNode
	for i := range engine.renders {
		for j := range engine.renders[i].zone {
			res = append(res, &engine.renders[i].zone[j])
		}
	}
	return res
}

// 集群已经运行的时间
func (engine *Engine) UpTime() time.Duration {
	return time.Duration(engine.UpdateCount) * DeltaT
}

func (engine *Engine) GetWorldTime() time.Time {
	return ZEROTIME.Add(engine.UpTime())
}
