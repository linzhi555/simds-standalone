package core

import (
	"log"
	"simds-standalone/config"
	"time"
)

// StateStorage 节点，用于共享状态的存储
type StateStorage struct {
	BasicNode
	LastSendTime time.Time
	Started      bool
	Schedulers   []string
	Workers      map[string]*NodeInfo
}

// NewStateStorage 创建新的StateStorage
func NewStateStorage(hostname string) *StateStorage {
	return &StateStorage{
		BasicNode: BasicNode{Host: hostname},
		Workers:   make(map[string]*NodeInfo),
	}
}

// StateCopy 复制一份集群状态拷贝
func (s *StateStorage) StateCopy() Vec[NodeInfo] {
	nodes := make(Vec[NodeInfo], 0, len(s.Workers))
	for _, ni := range s.Workers {
		nodes = append(nodes, *ni)
	}
	return nodes
}

func (s *StateStorage) Debug() { log.Println(s.Workers) }

func (s *StateStorage) Update(msg Message) {

	switch msg.Content {

	case "SignalBoot":
		s.LastSendTime = s.Os.GetTime()
		s.Os.Run(func() {
			for {
				time.Sleep(time.Duration(config.Val.StateUpdatePeriod) * time.Millisecond)
				newMessage := Message{
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
		task := msg.Body.(TaskInfo)
		if s.Workers[task.Worker].CanAllocate(task.CpuRequest, task.MemoryRequest) {
			s.Workers[task.Worker].AddAllocated(task.CpuRequest, task.MemoryRequest)
			err := s.Os.Send(Message{
				From:    s.GetHostName(),
				To:      task.Worker,
				Content: "TaskRun",
				Body:    task,
			})
			if err != nil {
				panic(err)
			}

		} else {
			err := s.Os.Send(Message{
				From:    s.GetHostName(),
				To:      msg.From,
				Content: "TaskCommitFail",
				Body:    task,
			})
			if err != nil {
				panic(err)
			}
			err = s.Os.Send(Message{
				From:    s.GetHostName(),
				To:      msg.From,
				Content: "NodeInfosUpdate",
				Body:    Vec[NodeInfo]{*s.Workers[task.Worker]},
			})
			if err != nil {
				panic(err)
			}
		}
	case "TaskFinish":
		taskInfo := msg.Body.(TaskInfo)
		s.Workers[msg.From].SubAllocated(taskInfo.CpuRequest, taskInfo.MemoryRequest)
	case "NeedUpdateNodeInfo":
		s.LastSendTime = s.Os.GetTime()
		stateCopy := s.StateCopy()
		for _, scheduler := range s.Schedulers {
			err := s.Os.Send(Message{
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
		task := msg.Body.(TaskInfo)
		newMessage := Message{
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
		newMessage := Message{
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
