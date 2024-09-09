package engine

import (
	"fmt"
	"time"

	"simds-standalone/cluster/base"
	"simds-standalone/common"
	"simds-standalone/config"
	"simds-standalone/tracing/rules"
)

// MockNetwork 模拟的网络组件
type VirtualNetwork struct {
	NetLatency int32
	Waittings  common.Vec[base.Message]
	Ins        map[string]*common.Vec[base.Message]
	Outs       map[string]*common.Vec[base.Message]
}

func newVirtualNetwork() VirtualNetwork {
	return VirtualNetwork{
		NetLatency: config.Val.NetLatency,
		Waittings:  common.Vec[base.Message]{},
		Ins:        make(map[string]*common.Vec[base.Message]),
		Outs:       make(map[string]*common.Vec[base.Message]),
	}
}

// 对集群引擎的虚拟网络进行更新
func (network *VirtualNetwork) updateNetwork(tNow time.Time) {

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

			rules.CheckRulesThenExec(rules.RecvRules, tNow, &m)
		} else {
			network.Waittings[i].LeftTime -= DeltaT
		}
		if needDelete {
			network.Waittings.Delete(i)
		} else {
			i++
		}
	}
}
