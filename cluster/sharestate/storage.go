package sharestate

import (
	"log"
	"time"

	"simds-standalone/cluster/base"
	"simds-standalone/config"
)

// StateStorage 节点，用于共享状态的存储
type StateStorage struct {
	base.BasicActor
	LastSendTime time.Time
	Started      bool
	Schedulers   []string
	Workers      map[string]*base.NodeInfo
}

// NewStateStorage 创建新的StateStorage
func NewStateStorage(hostname string) *StateStorage {
	return &StateStorage{
		BasicActor: base.BasicActor{Host: hostname},
		Workers:    make(map[string]*base.NodeInfo),
	}
}

// StateCopy 复制一份集群状态拷贝
func (s *StateStorage) StateCopy() base.Vec[base.NodeInfo] {
	nodes := make(base.Vec[base.NodeInfo], 0, len(s.Workers))
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
					From:    s.GetHostName(),
					To:      s.GetHostName(),
					Head: "NeedUpdateNodeInfo",
				}
				err := s.Os.Send(newMessage)
				if err != nil {
					log.Println(err)
				}
			}
		})

	case "TaskRun":
		task := msg.Body.(base.TaskInfo)
		if s.Workers[task.Worker].CanAllocate(task.CpuRequest, task.MemoryRequest) {
			s.Workers[task.Worker].AddAllocated(task.CpuRequest, task.MemoryRequest)
			err := s.Os.Send(base.Message{
				From:    s.GetHostName(),
				To:      task.Worker,
				Head: "TaskRun",
				Body:    task,
			})
			if err != nil {
				panic(err)
			}

		} else {
			err := s.Os.Send(base.Message{
				From:    s.GetHostName(),
				To:      msg.From,
				Head: "TaskCommitFail",
				Body:    task,
			})
			if err != nil {
				panic(err)
			}
			err = s.Os.Send(base.Message{
				From:    s.GetHostName(),
				To:      msg.From,
				Head: "NodeInfosUpdate",
				Body:    base.Vec[base.NodeInfo]{*s.Workers[task.Worker]},
			})
			if err != nil {
				panic(err)
			}
		}
	case "TaskFinish":
		taskInfo := msg.Body.(base.TaskInfo)
		s.Workers[msg.From].SubAllocated(taskInfo.CpuRequest, taskInfo.MemoryRequest)
	case "NeedUpdateNodeInfo":
		s.LastSendTime = s.Os.GetTime()
		stateCopy := s.StateCopy()
		for _, scheduler := range s.Schedulers {
			err := s.Os.Send(base.Message{
				From:    s.GetHostName(),
				To:      scheduler,
				Head: "NodeInfosUpdate",
				Body:    *stateCopy.Clone(),
			})
			if err != nil {
				panic(err)
			}
		}

	case "TaskCommitFail":
		task := msg.Body.(base.TaskInfo)
		newMessage := base.Message{
			From:    s.GetHostName(),
			To:      msg.From,
			Head: "TaskDispense",
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
		newMessage := base.Message{
			From:    s.GetHostName(),
			To:      s.GetHostName(),
			Head: "NeedUpdateNodeInfo",
		}
		err := s.Os.Send(newMessage)
		if err != nil {
			panic(err)
		}
	}
}
