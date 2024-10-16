package sharestate

import (
	"log"
	"time"

	"simds-standalone/cluster/base"
	"simds-standalone/cluster/lib"
	"simds-standalone/config"
)

// StateStorage 节点，用于共享状态的存储
type StateStorage struct {
	base.BasicActor
	Schedulers   []string
	Workers      map[string]*lib.NodeInfo
}

// NewStateStorage 创建新的StateStorage
func NewStateStorage(hostname string) *StateStorage {
	return &StateStorage{
		BasicActor: base.BasicActor{Host: hostname},
		Workers:    make(map[string]*lib.NodeInfo),
	}
}

// StateCopy 复制一份集群状态拷贝
func (s *StateStorage) StateCopy() lib.VecNodeInfo {
	nodes := make(lib.VecNodeInfo, 0, len(s.Workers))
	for _, ni := range s.Workers {
		nodes = append(nodes, *ni)
	}
	return nodes
}

func (s *StateStorage) Debug() { log.Println(s.Workers) }

func (s *StateStorage) Update(msg base.Message) {

	switch msg.Head {

	case "SignalBoot":
		s.Os.SetInterval(func() {
			s.Os.Send(base.Message{
				From: s.GetAddress(),
				To:   s.GetAddress(),
				Head: "SignalUpdate",
			})
		}, time.Duration(config.Val.StateUpdatePeriod)*time.Millisecond)

	case "TaskRun":
		task := msg.Body.(lib.TaskInfo)
		if s.Workers[task.Worker].CanAllocate(task.CpuRequest, task.MemoryRequest) {
			s.Workers[task.Worker].AddAllocated(task.CpuRequest, task.MemoryRequest)
			s.Os.Send(base.Message{
				From: s.GetAddress(),
				To:   task.Worker,
				Head: "TaskRun",
				Body: task,
			})

		} else {
			s.Os.Send(base.Message{
				From: s.GetAddress(),
				To:   msg.From,
				Head: "VecNodeInfoUpdate",
				Body: lib.VecNodeInfo{*s.Workers[task.Worker]},
			})

			s.Os.Send(base.Message{
				From: s.GetAddress(),
				To:   msg.From,
				Head: "TaskCommitFail",
				Body: task,
			})

		}
	case "TaskFinish":
		taskInfo := msg.Body.(lib.TaskInfo)
		s.Workers[msg.From].SubAllocated(taskInfo.CpuRequest, taskInfo.MemoryRequest)
	case "SignalUpdate":
		for _, scheduler := range s.Schedulers {
			s.Os.Send(base.Message{
				From: s.GetAddress(),
				To:   scheduler,
				Head: "VecNodeInfoUpdate",
				Body: s.StateCopy(),
			})
		}
	case "TaskCommitFail":
		task := msg.Body.(lib.TaskInfo)
		newMessage := base.Message{
			From: s.GetAddress(),
			To:   msg.From,
			Head: "TaskDispense",
			Body: task,
		}
		s.Os.Send(newMessage)

	}

}
