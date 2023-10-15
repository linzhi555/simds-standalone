package main

import (
	"time"
)

// 组件名定义
const (
	CMockNetWork   ComponentName = "MockNetwork"
	CTaskGen       ComponentName = "TaskGen"
	CScheduler     ComponentName = "Scheduler"
	CResouceManger ComponentName = "ResourceManager"
	CStateStorage  ComponentName = "StateStorage"
)

// OsApi 系统调用 抽象接口
type OsApi interface {
	GetTime() time.Time
	Net() NetInterface
}

// NodeComponent 节点组件抽象
type NodeComponent interface {
	Component
	SetOsApi(OsApi)
}

// TaskGen 组件
type TaskGen struct {
	Os        OsApi
	Host      string
	StartTime time.Time
	CurTaskId int
	Receivers []string
}

// NewTaskGen 创造空的TaskGen
func NewTaskGen(hostname string) *TaskGen {
	return &TaskGen{
		Host:      hostname,
		CurTaskId: 0,
	}
}

// Component For NodeComponent interface
func (n TaskGen) Component() ComponentName { return CTaskGen }

// SetOsApi for NodeComponent interface
func (n *TaskGen) SetOsApi(osapi OsApi) { n.Os = osapi }

// Scheduler 组件
type Scheduler struct {
	Os           OsApi
	Host         string
	Workers      map[string]*NodeInfo
	LocalNode    *NodeInfo
	WaitSchedule Vec[TaskInfo]
	TasksStatus  map[string]*TaskInfo
}

// NewScheduler 创造新的Scheduler
func NewScheduler(hostname string) *Scheduler {
	return &Scheduler{
		Host:         hostname,
		Workers:      make(map[string]*NodeInfo),
		WaitSchedule: Vec[TaskInfo]{},
		TasksStatus:  make(map[string]*TaskInfo),
	}
}

// Component For NodeComponent interface
func (s Scheduler) Component() ComponentName { return CScheduler }

// SetOsApi For NodeComponent interface
func (s *Scheduler) SetOsApi(osapi OsApi) { s.Os = osapi }

// GetAllWokersName 返回worker 名称列表
func (s *Scheduler) GetAllWokersName() []string {
	keys := make([]string, 0, len(s.Workers))
	for k := range s.Workers {
		keys = append(keys, k)
	}
	return keys
}

// StateStorage 组件，用于共享状态的存储
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

// Component For NodeComponent interface
func (s StateStorage) Component() ComponentName { return CStateStorage }

// SetOsApi For NodeComponent interface
func (s *StateStorage) SetOsApi(osapi OsApi) { s.Os = osapi }

// ResourceManager 组件
type ResourceManager struct {
	Os                 OsApi
	Host               string
	Tasks              map[string]*TaskInfo
	Node               NodeInfo
	TaskFinishReceiver string // if it is not zero , the receiver wiil get the notifiction
}

// NewResourceManager 创建新的ResourceManager
func NewResourceManager(host string) *ResourceManager {
	return &ResourceManager{
		Host:  host,
		Tasks: make(map[string]*TaskInfo),
	}
}

// Component For NodeComponent interface
func (n ResourceManager) Component() ComponentName { return CResouceManger }

// SetOsApi For NodeComponent interface
func (n *ResourceManager) SetOsApi(osapi OsApi) { n.Os = osapi }
