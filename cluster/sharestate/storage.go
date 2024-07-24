package sharestate

import (
	"log"
	"simds-standalone/config"
	"simds-standalone/core"
	"time"
)

// StateStorage 节点，用于共享状态的存储
type StateStorage struct {
	core.BasicNode
	LastSendTime time.Time
	Started      bool
	Schedulers   []string
	Workers      map[string]*core.NodeInfo
}

// NewStateStorage 创建新的StateStorage
func NewStateStorage(hostname string) *StateStorage {
	return &StateStorage{
		BasicNode: core.BasicNode{Host: hostname},
		Workers:   make(map[string]*core.NodeInfo),
	}
}

// StateCopy 复制一份集群状态拷贝
func (s *StateStorage) StateCopy() core.Vec[core.NodeInfo] {
	nodes := make(core.Vec[core.NodeInfo], 0, len(s.Workers))
	for _, ni := range s.Workers {
		nodes = append(nodes, *ni)
	}
	return nodes
}

func (s *StateStorage) Debug() { log.Println(s.Workers) }

func (s *StateStorage) Update(msg core.Message) {

	switch msg.Content {

	case "SignalBoot":
		s.LastSendTime = s.Os.GetTime()
		s.Os.Run(func() {
			for {
				time.Sleep(time.Duration(config.Val.StateUpdatePeriod) * time.Millisecond)
				newMessage := core.Message{
					From:    s.GetHostName(),
					To:      s.GetHostName(),
					Content: "NeedUpdateNodeInfo",
				}
				err := s.Os.Send(newMessage)
				if err != nil {
					log.Println(err)
				}
			}
		})

	case "TaskRun":
		task := msg.Body.(core.TaskInfo)
		if s.Workers[task.Worker].CanAllocate(task.CpuRequest, task.MemoryRequest) {
			s.Workers[task.Worker].AddAllocated(task.CpuRequest, task.MemoryRequest)
			err := s.Os.Send(core.Message{
				From:    s.GetHostName(),
				To:      task.Worker,
				Content: "TaskRun",
				Body:    task,
			})
			if err != nil {
				panic(err)
			}

		} else {
			err := s.Os.Send(core.Message{
				From:    s.GetHostName(),
				To:      msg.From,
				Content: "TaskCommitFail",
				Body:    task,
			})
			if err != nil {
				panic(err)
			}
			err = s.Os.Send(core.Message{
				From:    s.GetHostName(),
				To:      msg.From,
				Content: "NodeInfosUpdate",
				Body:    core.Vec[core.NodeInfo]{*s.Workers[task.Worker]},
			})
			if err != nil {
				panic(err)
			}
		}
	case "TaskFinish":
		taskInfo := msg.Body.(core.TaskInfo)
		s.Workers[msg.From].SubAllocated(taskInfo.CpuRequest, taskInfo.MemoryRequest)
	case "NeedUpdateNodeInfo":
		s.LastSendTime = s.Os.GetTime()
		stateCopy := s.StateCopy()
		for _, scheduler := range s.Schedulers {
			err := s.Os.Send(core.Message{
				From:    s.GetHostName(),
				To:      scheduler,
				Content: "NodeInfosUpdate",
				Body:    *stateCopy.Clone(),
			})
			if err != nil {
				panic(err)
			}
		}

	case "TaskCommitFail":
		task := msg.Body.(core.TaskInfo)
		newMessage := core.Message{
			From:    s.GetHostName(),
			To:      msg.From,
			Content: "TaskDispense",
			Body:    task,
		}
		err := s.Os.Send(newMessage)
		if err != nil {
			panic(err)
		}

	}

}

func (s *StateStorage) SimulateTasksUpdate() {
	if s.Os.GetTime().Sub(s.LastSendTime).Milliseconds() > int64(config.Val.StateUpdatePeriod) {
		newMessage := core.Message{
			From:    s.GetHostName(),
			To:      s.GetHostName(),
			Content: "NeedUpdateNodeInfo",
		}
		err := s.Os.Send(newMessage)
		s.Os.LogInfo("stdout", s.GetHostName(), newMessage.Content)
		if err != nil {
			panic(err)
		}
	}
}
