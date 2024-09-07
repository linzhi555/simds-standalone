package engine

import (
	"time"

	"simds-standalone/cluster/base"
	"simds-standalone/tracing/rules"
)

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
	return float32(DeltaT) / float32(self.Difficulty)
}

func NewVirtualNode(engine *Engine, actors ...base.Actor) *VirtualNode {
	var vnode VirtualNode
	vnode.engine = engine
	vnode.updatefunc = _defaultUpdateFunc
	vnode.actors = make(map[string]*EngineActor)
	for _, actor := range actors {
		vnode.AddActor(actor)
	}
	return &vnode
}

func (vnode *VirtualNode) AddActor(actor base.Actor) {

	vnode.actors[actor.GetAddress()] = &EngineActor{actor, ActorHideStatus{}}
}

func (vnode *VirtualNode) Update() {
	var lastState []ActorHideStatus
	for _, actor := range vnode.actors {
		lastState = append(lastState, actor.hide)
	}

	for _, actor := range vnode.actors {

		switch v := actor.model.(type) {
		case *timer:
			simulateTimer(v, &actor.hide)
			if actor.hide.NeedDestroy {
				delete(vnode.actors, v.Host)
			}
			continue
		case *cmdExecutor:
			simulateExecutor(v, &actor.hide)
			if actor.hide.NeedDestroy {
				delete(vnode.actors, v.Host)
			}
			continue
		}

		if actor.hide.IsBusy {
			actor.UpdateProgress(vnode.engine.GetWorldTime(), vnode.updatefunc(lastState, actor.hide))
		} else if msg, err := vnode.engine.Network.Outs[actor.model.GetAddress()].Dequeue(); err == nil {
			t := time.Now()

			actor.model.Update(msg)   // update the data status of the actor
			costTime := time.Since(t) // record the time cost, the value is treat as the task's "difficulty"
			rules.CheckRulesThenExec(rules.MsgDealRules, vnode.engine.GetWorldTime(), &msg)

			actor.hide.ToBusy(&msg, costTime)
		}
	}
}
