package engine

import (
	"simds-standalone/cluster/base"
	"simds-standalone/config"
	"simds-standalone/tracing/rules"
	"time"
)

// ZEROTIME 模拟开始的现实时间，以此作为模拟器的零点时间
var ZEROTIME time.Time = time.Now()

// 每次更新代表的时间长度
var DeltaT time.Duration = time.Second / time.Duration(config.Val.FPS)

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
