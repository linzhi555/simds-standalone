package engine

import (
	"fmt"
	"simds-standalone/cluster/base"
	"simds-standalone/common"
	"simds-standalone/tracing/rules"
	"strings"
	"time"
)

// timer and emd cmdExecutor 是特殊的actor
type timer struct {
	base.BasicActor
	isRepeat bool
	timeOut  time.Duration
	callback func()
}

func (*timer) Update(base.Message) {}

func simulateTimer(t *timer, hide *ActorHideStatus) {
	if hide.NeedDestroy {
		return
	}

	hide.Progress.Add(float32(DeltaT) / float32(t.timeOut))
	if hide.Progress.IsFinished() {
		t.callback()
		if t.isRepeat {
			hide.Progress = 0
		} else {
			hide.NeedDestroy = true
		}
	}
}

type cmdExecutor struct {
	base.BasicActor
	cmd      string
	lifetime time.Duration
	callback func(error)
}

func (*cmdExecutor) Update(base.Message) {}

func newCmdExecutor(host string, f func(error), cmd string) *cmdExecutor {
	c := &cmdExecutor{
		BasicActor: base.BasicActor{
			Host: host + "_cmd_" + cmd,
		},
		cmd:      cmd,
		callback: f,
	}

	if res, ok := strings.CutPrefix(c.cmd, "sleep "); ok {
		c.lifetime = time.Duration(common.Str_to_float64(res)*1000000) * time.Microsecond
	}
	return c
}

func simulateExecutor(cmd *cmdExecutor, hide *ActorHideStatus) {
	if hide.NeedDestroy {
		return
	}

	hide.Progress.Add(float32(DeltaT) / float32(cmd.lifetime))

	if hide.Progress.IsFinished() {
		cmd.callback(nil)
		hide.Progress = 0
		hide.NeedDestroy = true
	}
}

// MockOs 为组件提供模拟的系统调用
type EngineOs struct {
	addr   string
	node   *VirtualNode
	engine *Engine
	Out    *common.Vec[base.Message]
	In     *common.Vec[base.Message]
}

// GetTime 提供模拟时间
func (o *EngineOs) GetTime() time.Time {
	return o.engine.GetWorldTime()
}

func (o *EngineOs) SetInterval(callback func(), t time.Duration) {
	o.node.AddActor(&timer{
		BasicActor: base.BasicActor{
			Host: o.addr + "rt" + fmt.Sprint(t) + common.GenerateUID()[0:4], // ot mean: repeat timer
		},
		isRepeat: true,
		timeOut:  t,
		callback: callback,
	})
}

func (o *EngineOs) SetTimeOut(callback func(), t time.Duration) {
	o.node.AddActor(&timer{
		BasicActor: base.BasicActor{
			Host: o.addr + "ot" + fmt.Sprint(t) + common.GenerateUID()[0:4], // ot mean: once timer
		},
		isRepeat: false,
		timeOut:  t,
		callback: callback,
	})
}

func (o *EngineOs) RunCmd(callback func(err error), cmd string) {
	o.node.AddActor(newCmdExecutor(o.addr, callback, cmd))
}

func (o *EngineOs) Send(m base.Message) {
	m.Id = string(common.GenerateUID())
	o.Out.InQueueBack(m)
	rules.CheckRulesThenExec(rules.SendRules, o.GetTime(), &m)
}
