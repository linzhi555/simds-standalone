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
	UpdateCount uint64

	ctx      context.Context
	stopFunc func()

	NodeUpdateCost time.Duration
	renders        []renderThread

	NetUpdateCost time.Duration
	Network       VirtualNetwork
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
			node := *NewVirtualNode(&e, cluster.Nodes[j].Actors...)
			nodes = append(nodes, node)
		}

		e.renders = append(e.renders, renderThread{
			ctx:  e.ctx,
			zone: nodes,
		})
	}

	// init actor os
	for _, node := range e.nodes() {
		for _, actor := range node.actors {

			e.Network.Ins[actor.model.GetAddress()] = &common.Vec[base.Message]{}
			e.Network.Outs[actor.model.GetAddress()] = &common.Vec[base.Message]{}

			os := EngineOs{
				addr:   actor.model.GetAddress(),
				engine: &e,
				node:   node,
			}
			os.Send(base.Message{
				From: os.addr,
				To:   os.addr,
				Head: "SignalBoot",
				Body: lib.Signal("SignalBoot"),
			})
			actor.model.SetOsApi(&os)
		}
	}

	// start render thread
	for i := range e.renders {
		e.renders[i].goWorking()
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

// 对集群更新一次状态，推进一个单位时间
func (engine *Engine) Update() {
	engine.updateNodes()
	engine.Network.updateNetwork(engine.GetWorldTime())
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
