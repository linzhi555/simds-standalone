package lib

import (
	"fmt"

	"simds-standalone/cluster/base"
	"simds-standalone/common"
)

// Scheduler 组件
type CenterScheduler struct {
	base.BasicActor
	TaskMap map[string]*TaskInfo
	Workers map[string]*NodeInfo
	Storage string
}

// NewCenterScheduler 创造新的Scheduler
func NewCenterScheduler(hostname string) *CenterScheduler {
	return &CenterScheduler{
		BasicActor: base.BasicActor{Host: hostname},
		Workers:    make(map[string]*NodeInfo),
	}
}

func (s *CenterScheduler) Debug() {
	fmt.Println("task queues:")
	fmt.Println()
	fmt.Println("task status:")
}

func (s *CenterScheduler) Update(msg base.Message) {
	switch msg.Head {

	case "TaskDispense", "TaskCommitFail":
		task := msg.Body.(TaskInfo)
		dstWorker, ok := schdulingAlgorithm(s, &task)
		if ok {
			task.Worker = dstWorker
			s.Workers[task.Worker].AddAllocated(task.CpuRequest, task.MemoryRequest)
			receiver := task.Worker
			if s.Storage != "" {
				receiver = s.Storage
			}
			task.Worker = dstWorker
			err := s.Os.Send(base.Message{
				From: s.GetAddress(),
				To:   receiver,
				Head: "TaskRun",
				Body: task,
			})
			if err != nil {
				panic(err)
			}
		} else {
			err := s.Os.Send(base.Message{
				From: s.GetAddress(),
				To:   msg.From,
				Head: "TaskScheduleFail",
				Body: task,
			})
			if err != nil {
				panic(err)
			}
		}
	case "TaskFinish":
		taskInfo := msg.Body.(TaskInfo)
		s.Workers[msg.From].SubAllocated(taskInfo.CpuRequest, taskInfo.MemoryRequest)

	case "VecNodeInfoUpdate":
		nodeinfoList := msg.Body.(VecNodeInfo)
		for _, ni := range nodeinfoList {
			s.Workers[ni.Addr] = ni.Clone()
		}
	default:

	}
}


// schdulingAlgorithm 简单的首次适应调度算法，找到第一个合适调度的节点,找不到 ok返回false
func schdulingAlgorithm(scheduler *CenterScheduler, task *TaskInfo) (dstAddr string, ok bool) {
	dstAddr = ""

	keys := make([]string, 0, len(scheduler.Workers))
	for k := range scheduler.Workers {
		keys = append(keys, k)
	}
	common.ShuffleStringSlice(keys)
	for _, workerAddr := range keys {
		nodeinfo := scheduler.Workers[workerAddr]
		if nodeinfo.CanAllocate(task.CpuRequest, task.MemoryRequest) {
			dstAddr = workerAddr
		}
	}

	if dstAddr == "" {
		return dstAddr, false
	}
	return dstAddr, true
}
