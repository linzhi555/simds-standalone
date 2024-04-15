package core

import (
	"fmt"
	"math/rand"
	"simds-standalone/common"
	"simds-standalone/config"
)

// Scheduler 组件
type CenterScheduler struct {
	Os           OsApi
	Host         string
	Workers      map[string]*NodeInfo
	LocalNode    *NodeInfo
	WaitSchedule Vec[TaskInfo]
	TasksStatus  map[string]*TaskInfo
}

// NewCenterScheduler 创造新的Scheduler
func NewCenterScheduler(hostname string) *CenterScheduler {
	return &CenterScheduler{
		Host:         hostname,
		Workers:      make(map[string]*NodeInfo),
		WaitSchedule: Vec[TaskInfo]{},
		TasksStatus:  make(map[string]*TaskInfo),
	}
}

// SetOsApi For NodeComponent interface
func (s *CenterScheduler) SetOsApi(osapi OsApi) { s.Os = osapi }

func (s *CenterScheduler) Debug() {
	fmt.Println("task queues:")
	for _, task := range s.WaitSchedule {
		fmt.Printf("%+v\n", task)
	}

	fmt.Println()
	fmt.Println("task status:")
	for task, state := range s.TasksStatus {
		fmt.Printf("task:%+v status:%+v \n", task, state)
	}

}

// GetAllWokersName 返回worker 名称列表
func (s *CenterScheduler) GetAllWokersName() []string {
	keys := make([]string, 0, len(s.Workers))
	for k := range s.Workers {
		keys = append(keys, k)
	}
	return keys
}

// CenterSchedulerUpdate 模拟器每次tick时对中心化集群的调度器组件进行初始化
// 从网络中取出消息进行处理，然后进行有次数限制的调度动作
func (s *CenterScheduler) Update() {

	for !s.Os.Net().Empty() {
		newMessage, err := s.Os.Net().Recv()
		if err != nil {
			panic(err)
		}

		if newMessage.Content == "TaskDispense" {
			task := newMessage.Body.(TaskInfo)
			task.Status = "WaitSchedule"
			s.WaitSchedule.InQueue(task)
			s.Os.LogInfo(s.Os, "received TaskDispense", task)
		}

		if newMessage.Content == "TaskFinish" {
			taskInfo := newMessage.Body.(TaskInfo)
			s.Workers[newMessage.From].SubAllocated(taskInfo.CpuRequest, taskInfo.MemoryRequest)
			s.Os.LogInfo(s.Os, "received TaskFinish", newMessage.From, taskInfo)
		}

	}

	var maxScheduleTime = schdulingAlgorithmTimes(config.Val.SchedulerPerformance)
	for i := 0; i < maxScheduleTime; i++ {

		task, err := s.WaitSchedule.Dequeue()
		if err != nil {
			break
		}

		dstWorker, ok := schdulingAlgorithm(s, &task)
		if ok {
			task.Worker = dstWorker
			task.Status = "Allocated"
			s.Workers[task.Worker].AddAllocated(task.CpuRequest, task.MemoryRequest)
			newMessage := Message{
				From:    s.Os.Net().GetAddr(),
				To:      task.Worker,
				Content: "TaskRun",
				Body:    task,
			}
			err := s.Os.Net().Send(newMessage)
			if err != nil {
				panic(err)
			}

			s.Os.LogInfo(s.Os, "sendtask to", task.Worker, task)
		} else {
			s.WaitSchedule.InQueueFront(task)

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
