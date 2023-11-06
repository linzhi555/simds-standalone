package main

import (
	"fmt"
	"simds-standalone/common"
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
	Src       []SrcNode
}

type SrcNode struct {
	time time.Duration
	task TaskInfo
}

// var TaskSrcGenerate func() []SrcNode = noWaveTaskStream
var TaskSrcGenerate func() []SrcNode = onePeakTaskStream

// NewTaskGen 创造空的TaskGen
func NewTaskGen(hostname string) *TaskGen {

	taskgen := &TaskGen{
		Host:      hostname,
		CurTaskId: 0,
	}

	taskgen.Src = TaskSrcGenerate()

	return taskgen
}

// 负载没有波动的连续任务流
func noWaveTaskStream() []SrcNode {
	taskNumPerSecond := Config.TaskNumFactor * float32(Config.NodeNum)
	allTasksNum := int(10 * taskNumPerSecond)
	src := make([]SrcNode, 0, allTasksNum)
	for i := 0; i < allTasksNum; i++ {
		newTask := TaskInfo{
			Id:            fmt.Sprintf("task%d", i),
			CpuRequest:    common.RandIntWithRange(Config.TaskCpu, 0.5),
			MemoryRequest: common.RandIntWithRange(Config.TaskMemory, 0.5),
			LifeTime:      time.Duration(common.RandIntWithRange(Config.TaskLifeTime, 0.5)) * time.Millisecond,
			Status:        "submit",
		}

		t := time.Duration(int64(i) * 10 * int64(time.Second) / int64(allTasksNum))

		src = append(src, SrcNode{t, newTask})

	}
	return src
}

func pow2(x int64) int64 {
	return x * x
}

// 有一个峰值的连续任务流
func onePeakTaskStream() []SrcNode {
	taskNumPerSecond := Config.TaskNumFactor * float32(Config.NodeNum)
	allTasksNum := int(10 * taskNumPerSecond)
	src := make([]SrcNode, 0, allTasksNum)
	for i := 0; i < allTasksNum; i++ {
		newTask := TaskInfo{
			Id:            fmt.Sprintf("task%d", i),
			CpuRequest:    common.RandIntWithRange(Config.TaskCpu, 0.5),
			MemoryRequest: common.RandIntWithRange(Config.TaskMemory, 0.5),
			LifeTime:      time.Duration(common.RandIntWithRange(Config.TaskLifeTime, 0.5)) * time.Millisecond,
			Status:        "submit",
		}

		var t time.Duration
		if i <= allTasksNum/4 {
			t = time.Duration(int64(i)*10*int64(time.Second)/int64(allTasksNum)) * 3 / 2
		} else if i <= allTasksNum*3/4 {
			t = src[allTasksNum/4].time
			t += time.Duration(int64(i-(allTasksNum/4)) * 10 * int64(time.Second) / int64(allTasksNum) * 3 / 4)
		} else {
			t = src[allTasksNum*3/4].time
			t += time.Duration(int64(i-(allTasksNum*3/4)) * 10 * int64(time.Second) / int64(allTasksNum) * 3 / 2)
		}

		src = append(src, SrcNode{time.Duration(t), newTask})

	}
	return src
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
	Os    OsApi
	Host  string
	Tasks map[string]*TaskInfo
	//Node               NodeInfo // do not store the information , calculate when needed from tasks
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
