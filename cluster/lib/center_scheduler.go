package lib

import (
	"fmt"
	"log"
	"time"

	"simds-standalone/cluster/base"
	"simds-standalone/common"
	"simds-standalone/config"
)

type scheduleFunc func(workers map[string]*NodeInfo, task *TaskInfo) (string, bool)

// Common Scheduler Acotr
type CenterScheduler struct {
	base.BasicActor
	Workers       map[string]*NodeInfo // the worker's resource info
	WaittingTasks common.Vec[TaskInfo]
	Algorithm     scheduleFunc

	// the worker's Manager
	// normally the Worker is managed by himself
	// in some case .the worker is managed by third person, such as in sharestate cluster
	WorkerManager map[string]string
}

// NewCenterScheduler 创造新的Scheduler
func NewCenterScheduler(hostname string) *CenterScheduler {
	scheduler := CenterScheduler{
		BasicActor:    base.BasicActor{Host: hostname},
		Workers:       make(map[string]*NodeInfo),
		WorkerManager: make(map[string]string),
	}

	switch config.Val.ScheduleFunc {
	case "firstFit":
		scheduler.Algorithm = firstFit
	case "lowestCPU":
		scheduler.Algorithm = lowestCPU

	default:
		panic("please give me the right schedule func to initilaze the scheduler")

	}

	return &scheduler
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
			dstWorker, ok := s.Algorithm(s.Workers, task)
			if ok {
				task.Worker = dstWorker
				s.Workers[task.Worker].AddAllocated(task.CpuRequest, task.MemoryRequest)
				receiver, ok := s.WorkerManager[task.Worker]
				if !ok {
					receiver = task.Worker
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
			s.WorkerManager[ni.Addr] = msg.From
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

// firstFit 简单的首次适应调度算法，
// 找到了返回 目标worker,true
// 找不到返回 "",false
func firstFit(workers map[string]*NodeInfo, task *TaskInfo) (string, bool) {
	result := ""

	ids := make([]string, 0, len(workers))
	for k := range workers {
		ids = append(ids, k)
	}

	common.ShuffleStringSlice(ids)
	for _, id := range ids {
		nodeinfo := workers[id]
		if nodeinfo.CanAllocate(task.CpuRequest, task.MemoryRequest) {
			result = id
			break
		}
	}

	if result == "" {
		return result, false
	}

	return result, true
}

// 寻找CPU最低最空闲的的节点，
// 找到了返回 目标worker,true
// 找不到返回 "",false
func lowestCPU(workers map[string]*NodeInfo, task *TaskInfo) (string, bool) {
	result := ""

	ids := make([]string, 0, len(workers))
	for k := range workers {
		ids = append(ids, k)
	}

	common.ShuffleStringSlice(ids)

	lowestCPUWorker := ""
	lowestCPURercent := float32(1.0)

	for _, id := range ids {
		nodeinfo := workers[id]
		if nodeinfo.CanAllocate(task.CpuRequest, task.MemoryRequest) {
			workcpuPer := nodeinfo.CpuPercent()
			if workcpuPer < lowestCPURercent {
				lowestCPURercent = workcpuPer
				lowestCPUWorker = id
			}
		}
	}

	if lowestCPUWorker == "" {
		return result, false
	}

	return lowestCPUWorker, true
}
