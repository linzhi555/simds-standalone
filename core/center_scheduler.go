package core

import (
	"fmt"
	"simds-standalone/common"
)

// Scheduler 组件
type CenterScheduler struct {
	BasicNode
	TaskMap map[string]*TaskInfo
	Workers map[string]*NodeInfo
	storage string
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

		s.Os.LogInfo("stdout", s.GetHostName(), event.Content, fmt.Sprint(event.Body))

		switch event.Content {

		case "TaskDispense", "TaskCommitFail":
			task := event.Body.(TaskInfo)
			dstWorker, ok := schdulingAlgorithm(s, &task)
			if ok {
				task.Worker = dstWorker
				s.Workers[task.Worker].AddAllocated(task.CpuRequest, task.MemoryRequest)
				receiver := task.Worker
				if s.storage != "" {
					receiver = s.storage
				}
				task.Worker = dstWorker
				newMessage := Message{
					From:    s.GetHostName(),
					To:      receiver,
					Content: "TaskRun",
					Body:    task,
				}
				err := s.Os.Send(newMessage)
				if err != nil {
					panic(err)
				}
				s.Os.LogInfo("stdout", s.GetHostName(), "TaskScheduled", fmt.Sprint(task))
			} else {
				newMessage := Message{
					From:    s.GetHostName(),
					To:      event.From,
					Content: "TaskCommitFail",
					Body:    task,
				}
				err := s.Os.Send(newMessage)
				if err != nil {
					panic(err)
				}
				s.Os.LogInfo("stdout", s.GetHostName(), "TaskScheduledFail", fmt.Sprint(task))
			}
		case "TaskFinish":
			taskInfo := event.Body.(TaskInfo)
			s.Workers[event.From].SubAllocated(taskInfo.CpuRequest, taskInfo.MemoryRequest)

		case "NodeInfosUpdate":
			nodeinfoList := event.Body.(Vec[NodeInfo])
			for _, ni := range nodeinfoList {
				s.Workers[ni.Addr] = ni.Clone()
			}
			s.Os.LogInfo("stdout", s.GetHostName(), "NodeInfosUpdate")

			/* 		case "TaskScheduled":
			task := event.Body.(TaskInfo)
			s.Workers[task.Worker].AddAllocated(task.CpuRequest, task.MemoryRequest)

			receiver := task.Worker

			if s.storage != "" {
				receiver = s.storage
			}

			newMessage := Message{
				From:    s.GetHostName(),
				To:      receiver,
				Content: "TaskRun",
				Body:    task,
			}
			err := s.Os.Send(newMessage)
			if err != nil {
				panic(err)
			}
			s.Os.LogInfo("stdout", s.GetHostName(), "TaskRun", task.Id) */
		}
	}
}

func (s *CenterScheduler) SimulateTasksUpdate() {

}

// 在一个调度器中，每次更新执行调度算法的次数，该函数的影响参数是
// performance : 该机器的性能参数 unit tasks / second
// func schdulingAlgorithmTimes(performance float32) int {
// 	times_float := performance / float32(config.Val.FPS) // 每次更新相当于时间 1 / config.Val/FPS秒
// 	base := int(times_float)
// 	var times_int int
// 	if rand.Float32() < times_float-float32(base) {
// 		times_int = base + 1
// 	} else {
// 		times_int = base
// 	}
// 	return times_int
// }

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
