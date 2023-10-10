package main

import (
	"fmt"
	"log"
	"simds-standalone/common"
	"time"
)

var ZEROTIME time.Time = time.Now()

type MockOs struct {
	getTimeFunc func() time.Time
	net         NetInterface
}

func (o *MockOs) GetTime() time.Time {
	return o.getTimeFunc()
}
func (o *MockOs) Net() NetInterface {
	return o.net
}

func EcsRunCluster(cluster Cluster) {
	simulator := NewEcs()
	newNet := NewMockNetWork(Config.NetLatency * MiliSecond)
	getTimeFunc := func() time.Time { return ZEROTIME.Add(time.Duration(simulator.UpdateCount) * time.Millisecond) }
	card := CreateMockNetCard("network1" + ":" + string(CMockNetWork))
	card.JoinNetWork(newNet)
	newNet.SetOsApi(
		&MockOs{
			getTimeFunc,
			card,
		},
	)

	simulator.AddEntities("network1", newNet)
	simulator.AddSystem(SystemName(string(newNet.Component()+"_update")), func(e *ECS) {
		e.ApplyToAllComponent(CMockNetWork, networkTick)
	})

	for _, node := range cluster.Nodes {
		var inited []Component
		for _, c := range node.Components {

			card := CreateMockNetCard(node.Name + ":" + string(c.Component()))
			card.JoinNetWork(newNet)

			c.SetOsApi(
				&MockOs{
					getTimeFunc,
					card,
				},
			)

			inited = append(inited, c)
		}
		simulator.AddEntities(EntityName(node.Name), inited...)
	}

	for k, f := range cluster.Setups {
		simulator.AddSystem(SystemName(string(k)+"_setup"), covertFuncToSystem(k, f, true))
	}

	for k, f := range cluster.Updates {
		simulator.AddSystem(SystemName(string(k)+"_update"), covertFuncToSystem(k, f, false))
	}

	frameNum := 12000
	for i := 0; i < frameNum; i++ {
		log.Println("simluating", i, "/", frameNum)
		simulator.Update()

	}

}

func covertFuncToSystem(c ComponentName, f func(interface{}), isSetup bool) func(e *ECS) {
	return func(e *ECS) {
		if isSetup {
			if e.UpdateCount != 0 {
				return
			}
		}

		componetTick := func(ecs *ECS, e EntityName, comp Component) Component {
			switch t := comp.(type) {
			case *TaskGen:
				f(t)
				return t
			case *Scheduler:
				f(t)
				return t
			case *ResourceManager:
				f(t)
				return t
			}
			return comp
		}
		e.ApplyToAllComponent(c, componetTick)
	}
}

func networkTick(ecs *ECS, entity EntityName, comp Component) Component {
	n := comp.(*MockNetwork)

	for _, in := range n.Ins {
		for !in.Empty() {
			newM, err := in.Dequeue()
			// message body can not be pointer
			common.AssertTypeIsNotPointer(newM.Body)
			if common.IsSameHost(newM.To, newM.From) {
				newM.LeftTime = 0
			} else {
				newM.LeftTime = n.NetLatency
			}

			if err != nil {
				panic(err)
			}

			LogInfo(n.Os, ": new message waitting to be send", newM)
			n.Waittings.InQueue(newM)
		}

	}
	for i := 0; i < len(n.Waittings); {
		m := n.Waittings[i]
		needDelete := false
		if m.LeftTime == 0 {
			LogInfo(n.Os, ": new message sended", m)
			out, ok := n.Outs[m.To]
			if !ok {
				panic(m.To + ":net can not reach")
			}
			needDelete = true
			out.InQueue(m)
		} else {
			n.Waittings[i].LeftTime -= 1
		}
		if needDelete {
			n.Waittings.Delete(i)
		} else {
			i += 1
		}

	}

	return n

}

func LogInfo(osapi OsApi, ins ...interface{}) {
	s := fmt.Sprint(osapi.GetTime().Sub(ZEROTIME).Milliseconds(), " ", "Info", " ", osapi.Net().GetAddr(), " ")
	for _, item := range ins {
		s += fmt.Sprint(item, " ")
	}
	fmt.Println(s)
}
