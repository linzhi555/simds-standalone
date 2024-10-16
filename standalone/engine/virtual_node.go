package engine

import (
	"time"

	"simds-standalone/cluster/base"
	"simds-standalone/common"
	"simds-standalone/config"
	"simds-standalone/tracing/rules"
)

type ActorHideStatus struct {
	IsBusy      bool
	NeedDestroy bool
	Progress    Progress
	LastMsg     *base.Message
	Difficulty  time.Duration
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
	os    EngineOs
	model base.Actor
	hide  ActorHideStatus
}

type VirtualNode struct {
	engine     *Engine
	updatefunc func([]ActorHideStatus, ActorHideStatus) float32
	actors     common.Vec[EngineActor]
}

func _defaultUpdateFunc(_ []ActorHideStatus, self ActorHideStatus) float32 {
	return config.Val.PerformanceRate * float32(DeltaT) / float32(self.Difficulty)
}

func NewVirtualNode(engine *Engine, actors ...base.Actor) *VirtualNode {
	var vnode VirtualNode
	vnode.engine = engine
	vnode.updatefunc = _defaultUpdateFunc
	vnode.actors = make(common.Vec[EngineActor], 0, len(actors))
	for _, actor := range actors {
		vnode.AddActor(actor)
	}
	return &vnode
}

func (vnode *VirtualNode) AddActor(actor base.Actor) {
	vnode.actors.InQueueBack(EngineActor{model: actor, hide: ActorHideStatus{}})
}

func (vnode *VirtualNode) Update() {
	lastState := make([]ActorHideStatus, 0, len(vnode.actors))
	for _, actor := range vnode.actors {
		lastState = append(lastState, actor.hide)
	}

	for i := 0; i < len(vnode.actors); i++ {
		actor := &vnode.actors[i]

		switch v := actor.model.(type) {
		case *timer:
			simulateTimer(v, &actor.hide)
			continue
		case *cmdExecutor:
			simulateExecutor(v, &actor.hide)
			continue
		}

		if actor.hide.IsBusy {
			actor.UpdateProgress(vnode.engine.GetWorldTime(), vnode.updatefunc(lastState, actor.hide))
		} else if msg, err := actor.os.In.Dequeue(); err == nil {
			t := time.Now()

			actor.model.Update(msg)   // update the data status of the actor
			costTime := time.Since(t) // record the time cost, the value is treat as the task's "difficulty"
			rules.CheckRulesThenExec(rules.MsgDealRules, vnode.engine.GetWorldTime(), &msg)

			actor.hide.ToBusy(&msg, costTime)
			actor.UpdateProgress(vnode.engine.GetWorldTime(), vnode.updatefunc(lastState, actor.hide))
		}
	}

	needDeletes := []int{}
	for i, actor := range vnode.actors {
		if actor.hide.NeedDestroy {
			needDeletes = append(needDeletes, i)
		}
	}
	for i := len(needDeletes) - 1; i >= 0; i-- {
		vnode.actors.Delete(needDeletes[i])
	}
}
