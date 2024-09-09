package lib

import (
	"fmt"
	"log"
	"time"

	"simds-standalone/cluster/base"
	"simds-standalone/common"
)

// Scheduler 组件
type CenterScheduler struct {
	base.BasicActor
	Workers       map[string]*NodeInfo
	WaittingTasks common.Vec[TaskInfo]
	Storage       string
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

	case "SignalBoot":
		s._setNextSchdulerTimer()

	case "TaskDispense", "TaskCommitFail":
		task := msg.Body.(TaskInfo)
		s.WaittingTasks.InQueueBack(task)

	case "SignalSchedule":

		for s.WaittingTasks.Len() > 0 {
			task := &s.WaittingTasks[0]
			dstWorker, ok := schdulingAlgorithm(s, task)
			if ok {
				task.Worker = dstWorker
				s.Workers[task.Worker].AddAllocated(task.CpuRequest, task.MemoryRequest)
				receiver := task.Worker
				if s.Storage != "" {
					receiver = s.Storage
				}
				task.Worker = dstWorker
				s.Os.Send(base.Message{
					From: s.GetAddress(),
					To:   receiver,
					Head: "TaskRun",
					Body: *(task.Clone()),
				})
				s.WaittingTasks.Delete(0)
			} else {
				break
			}
		}
		s._setNextSchdulerTimer()

	case "TaskFinish":

		taskInfo := msg.Body.(TaskInfo)
		s.Workers[msg.From].SubAllocated(taskInfo.CpuRequest, taskInfo.MemoryRequest)

	case "VecNodeInfoUpdate":
		nodeinfoList := msg.Body.(VecNodeInfo)
		for _, ni := range nodeinfoList {
			s.Workers[ni.Addr] = ni.Clone()
		}

	default:
		log.Panicln(msg)
	}
}

// 设置一个闹钟，提醒下一次检查一次任务队列
func (s *CenterScheduler) _setNextSchdulerTimer() {
	//设置一定时间检查任务队列一次。
	s.Os.SetTimeOut(func() {
		s.Os.Send(base.Message{
			From: s.GetAddress(),
			To:   s.GetAddress(),
			Head: "SignalSchedule",
			Body: Signal("SignalSchedule"),
		})
	}, 10*time.Millisecond)
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
