package core

import (
	"log"
	"time"
)

// StateStorage 节点，用于共享状态的存储
type StateStorage struct {
	Os           OsApi
	Host         string
	LastSendTime time.Time
	Workers      map[string]*NodeInfo
}

// NewStateStorage 创建新的StateStorage
func NewStateStorage(hostname string) *StateStorage {
	return &StateStorage{
		Host:    hostname,
		Workers: make(map[string]*NodeInfo),
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

// SetOsApi For NodeComponent interface
func (s *StateStorage) SetOsApi(osapi OsApi) { s.Os = osapi }

func (s *StateStorage) Debug() { log.Println(s.Workers) }
