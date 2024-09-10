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
			newM, _ := in.Pop()
			// message body can not be pointer
			common.AssertTypeIsNotPointer(newM.Body)
			if newM.To == newM.From {
				newM.LeftTime = 100 * time.Microsecond
			} else {
				newM.LeftTime = time.Duration(common.RandIntWithRange(network.NetLatency*1000, 0.3)) * time.Microsecond
				//newM.LeftTime = time.Duration(network.NetLatency) * 10 * time.Millisecond
			}

			network.Waittings.InQueueBack(newM)
		}
	}

	// 集中处理message vector比map[ActorId]Messge速度更快

	var deletes []int
	for i := 0; i < len(network.Waittings); i++ {
		m := &(network.Waittings[i])
		if m.LeftTime < 0 {
			out, ok := network.Outs[m.To]
			if !ok {
				panic(fmt.Sprint(m) + ":net can not reach")
			}
			out.InQueueBack(*m)
			deletes = append(deletes, i)
			rules.CheckRulesThenExec(rules.RecvRules, tNow, m)
		} else {
			network.Waittings[i].LeftTime -= DeltaT
		}

	}

	for i := len(deletes) - 1; i >= 0; i-- {
		network.Waittings.Delete(deletes[i])
	}

}
