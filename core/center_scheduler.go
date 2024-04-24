package core

import (
	"fmt"
	"math/rand"
	"simds-standalone/common"
	"simds-standalone/config"
)

// Scheduler 组件
type CenterScheduler struct {
	BasicNode
	TaskList Vec[TaskInfo]
	TaskMap  map[string]*TaskInfo
	Workers  map[string]*NodeInfo
}

// NewCenterScheduler 创造新的Scheduler
func NewCenterScheduler(hostname string) *CenterScheduler {
	return &CenterScheduler{
		BasicNode: BasicNode{Host: hostname},
		Workers:   make(map[string]*NodeInfo),
	}
}

func (s *CenterScheduler) Debug() {
	fmt.Println("task queues:")
	fmt.Println()
	fmt.Println("task status:")
}

// // GetAllWokersName 返回worker 名称列表
// func (s *CenterScheduler) GetAllWokersName() []string {
// 	keys := make([]string, 0, len(s.Workers))
// 	for k := range s.Workers {
// 		keys = append(keys, k)
// 	}
// 	return keys
// }

func (s *CenterScheduler) Update() {
	for s.Os.HasMessage() {
		event, err := s.Os.Recv()
		if err != nil {
			panic(err)
		}
		switch event.Content {
		case "TaskDispense":
			task := event.Body.(TaskInfo)
			task.Status = "WaitSchedule"
			s.TaskList.InQueue(task)
			s.Os.LogInfo("stdout", s.GetHostName(), "TaskDispense", fmt.Sprint(task))
		case "TaskFinish":
			taskInfo := event.Body.(TaskInfo)
			s.Workers[event.From].SubAllocated(taskInfo.CpuRequest, taskInfo.MemoryRequest)
			s.Os.LogInfo("stdout", s.GetHostName(), "TaskFinish", fmt.Sprint(taskInfo))
		case "TaskScheduled":
			task := event.Body.(TaskInfo)
			s.Workers[task.Worker].AddAllocated(task.CpuRequest, task.MemoryRequest)
			newMessage := Message{
				From:    s.GetHostName(),
				To:      task.Worker,
				Content: "TaskRun",
				Body:    task,
			}
			err := s.Os.Send(newMessage)
			if err != nil {
				panic(err)
			}
			s.Os.LogInfo("stdout", s.GetHostName(), "TaskRun", task.Id)

		}
	}
}

func (s *CenterScheduler) SimulateTasksUpdate() {
	var maxScheduleTime = schdulingAlgorithmTimes(config.Val.SchedulerPerformance)
	for i := 0; i < maxScheduleTime; i++ {
		task, err := s.TaskList.Dequeue()
		if err != nil {
			break
		}
		dstWorker, ok := schdulingAlgorithm(s, &task)
		if ok {

			task.Worker = dstWorker
			task.Status = "Allocated"
			newMessage := Message{
				From:    s.GetHostName(),
				To:      s.GetHostName(),
				Content: "TaskScheduled",
				Body:    task,
			}
			err := s.Os.Send(newMessage)
			if err != nil {
				panic(err)
			}
			s.Os.LogInfo("stdout", s.GetHostName(), "TaskScheduled", fmt.Sprint(task))
		} else {
			s.TaskList.InQueueFront(task)
		}

	}
}

// 在一个调度器中，每次更新执行调度算法的次数，该函数的影响参数是
// performance : 该机器的性能参数 unit tasks / second
func schdulingAlgorithmTimes(performance float32) int {
	times_float := performance / float32(config.Val.FPS) // 每次更新相当于时间 1 / config.Val/FPS秒
	base := int(times_float)
	var times_int int
	if rand.Float32() < times_float-float32(base) {
		times_int = base + 1
	} else {
		times_int = base
	}

	return times_int
}

// schdulingAlgorithm 简单的调度算法，找到第一个合适调度的节点,找不到 ok返回false
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
