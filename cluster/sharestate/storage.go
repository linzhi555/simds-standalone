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
	LastSendTime time.Time
	Started      bool
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
func (s *StateStorage) StateCopy() base.Vec[lib.NodeInfo] {
	nodes := make(base.Vec[lib.NodeInfo], 0, len(s.Workers))
	for _, ni := range s.Workers {
		nodes = append(nodes, *ni)
	}
	return nodes
}

func (s *StateStorage) Debug() { log.Println(s.Workers) }

func (s *StateStorage) Update(msg base.Message) {

	switch msg.Head {

	case "SignalBoot":
		s.LastSendTime = s.Os.GetTime()
		s.Os.Run(func() {
			for {
				time.Sleep(time.Duration(config.Val.StateUpdatePeriod) * time.Millisecond)
				newMessage := base.Message{
					From: s.GetHostName(),
					To:   s.GetHostName(),
					Head: "SignalUpdate",
				}
				err := s.Os.Send(newMessage)
				if err != nil {
					log.Println(err)
				}
			}
		})

	case "TaskRun":
		task := msg.Body.(lib.TaskInfo)
		if s.Workers[task.Worker].CanAllocate(task.CpuRequest, task.MemoryRequest) {
			s.Workers[task.Worker].AddAllocated(task.CpuRequest, task.MemoryRequest)
			err := s.Os.Send(base.Message{
				From: s.GetHostName(),
				To:   task.Worker,
				Head: "TaskRun",
				Body: task,
			})
			if err != nil {
				panic(err)
			}

		} else {
			err := s.Os.Send(base.Message{
				From: s.GetHostName(),
				To:   msg.From,
				Head: "VecNodeInfoUpdate",
				Body: base.Vec[lib.NodeInfo]{*s.Workers[task.Worker]},
			})
			if err != nil {
				panic(err)
			}
			err = s.Os.Send(base.Message{
				From: s.GetHostName(),
				To:   msg.From,
				Head: "TaskCommitFail",
				Body: task,
			})
			if err != nil {
				panic(err)
			}

		}
	case "TaskFinish":
		taskInfo := msg.Body.(lib.TaskInfo)
		s.Workers[msg.From].SubAllocated(taskInfo.CpuRequest, taskInfo.MemoryRequest)
	case "SignalUpdate":
		s.LastSendTime = s.Os.GetTime()
		stateCopy := s.StateCopy()
		for _, scheduler := range s.Schedulers {
			err := s.Os.Send(base.Message{
				From: s.GetHostName(),
				To:   scheduler,
				Head: "VecNodeInfoUpdate",
				Body: *stateCopy.Clone(),
			})
			if err != nil {
				panic(err)
			}
		}
	case "TaskCommitFail":
		task := msg.Body.(lib.TaskInfo)
		newMessage := base.Message{
			From: s.GetHostName(),
			To:   msg.From,
			Head: "TaskDispense",
			Body: task,
		}
		err := s.Os.Send(newMessage)
		if err != nil {
			panic(err)
		}

	}

}

func (s *StateStorage) SimulateTasksUpdate() {
	if s.Os.GetTime().Sub(s.LastSendTime).Milliseconds() > int64(config.Val.StateUpdatePeriod) {
		newMessage := base.Message{
			From: s.GetHostName(),
			To:   s.GetHostName(),
			Head: "SignalUpdate",
		}
		err := s.Os.Send(newMessage)
		if err != nil {
			panic(err)
		}
	}
}
