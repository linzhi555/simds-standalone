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
	NetLatency  time.Duration
	NetVariance float32
	Waittings   common.Vec[base.Message]
	Ins         map[string]*common.Vec[base.Message]
	Outs        map[string]*common.Vec[base.Message]
}

func newVirtualNetwork() VirtualNetwork {
	return VirtualNetwork{
		NetLatency:  time.Duration(config.Val.NetLatency*1000000) * time.Nanosecond,
		NetVariance: config.Val.NetLatencyVar,
		Waittings:   common.Vec[base.Message]{},
		Ins:         make(map[string]*common.Vec[base.Message]),
		Outs:        make(map[string]*common.Vec[base.Message]),
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
				variance := float64(network.NetVariance)

				newM.LeftTime = time.Duration(
					float64(network.NetLatency)*(1.0-variance) +
						common.ExponentialRand(float64(network.NetLatency)*variance),
				)

				//newM.LeftTime = network.NetLatency
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
			//network.Waittings[i].LeftTime -= time.Duration(common.ExponentialRand(float64(DeltaT))*rand.Float64()*2)
			network.Waittings[i].LeftTime -= DeltaT
		}

	}

	for i := len(deletes) - 1; i >= 0; i-- {
		network.Waittings.Delete(deletes[i])
	}

}
