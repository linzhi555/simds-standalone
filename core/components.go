package core

import (
	"fmt"
	"log"
	"simds-standalone/common"
	"simds-standalone/config"
	"time"
)

// 组件名定义
const (
	CMockNetWork   ComponentName = "MockNetwork"
	CTaskGen       ComponentName = "TaskGen"
	CScheduler     ComponentName = "Scheduler"
	CResouceManger ComponentName = "ResourceManager"
	CStateStorage  ComponentName = "StateStorage"
	CRaftManager   ComponentName = "RaftManager"
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

// NewTaskGen 创造空的TaskGen
func NewTaskGen(hostname string) *TaskGen {

	taskgen := &TaskGen{
		Host:      hostname,
		CurTaskId: 0,
	}

	switch config.Val.TaskMode {
	case "onePeak":
		taskgen.Src = onePeakTaskStream()
	case "noWave":
		taskgen.Src = noWaveTaskStream()
	case "trace":
		src := readTraceTaskStream(config.Val.TraceFile, 1.0, config.Val.SimulateDuration-10000)
		src = applyLoadRate(src, float64(config.Val.NodeNum)/float64(1000)*float64(config.Val.TaskNumFactor)/7.0)
		taskgen.Src = src
	default:
		panic("this mod is not implented")
	}

	return taskgen
}

// 负载没有波动的连续任务流
func noWaveTaskStream() []SrcNode {
	taskNumPerSecond := config.Val.TaskNumFactor * float32(config.Val.NodeNum)
	var sendDuration = time.Duration(config.Val.SimulateDuration-10000) * time.Millisecond
	allTasksNum := int(float32(sendDuration/time.Second) * taskNumPerSecond)
	src := make([]SrcNode, 0, allTasksNum)

	for i := 0; i < allTasksNum; i++ {
		newTask := TaskInfo{
			Id:            fmt.Sprintf("task%d", i),
			CpuRequest:    common.RandIntWithRange(config.Val.TaskCpu, 0.5),
			MemoryRequest: common.RandIntWithRange(config.Val.TaskMemory, 0.5),
			LifeTime:      time.Duration(common.RandIntWithRange(config.Val.TaskLifeTime, 0.5)) * time.Millisecond,
			Status:        "submit",
		}

		t := time.Duration(int64(i) * int64(sendDuration) / int64(allTasksNum))

		src = append(src, SrcNode{t, newTask})

	}
	return src
}

func pow2(x int64) int64 {
	return x * x
}

// 有一个峰值的连续任务流
func onePeakTaskStream() []SrcNode {
	taskNumPerSecond := config.Val.TaskNumFactor * float32(config.Val.NodeNum)
	baseTimeDelta := int64(time.Second) / int64(taskNumPerSecond)
	src := make([]SrcNode, 0)
	for i := 0; ; i++ {
		newTask := TaskInfo{
			Id:            fmt.Sprintf("task%d", i),
			CpuRequest:    common.RandIntWithRange(config.Val.TaskCpu, 0.5),
			MemoryRequest: common.RandIntWithRange(config.Val.TaskMemory, 0.5),
			LifeTime:      time.Duration(common.RandIntWithRange(config.Val.TaskLifeTime, 0.5)) * time.Millisecond,
			Status:        "submit",
		}

		var t time.Duration

		var sendDuration = time.Duration(config.Val.SimulateDuration-10000) * time.Millisecond

		if i == 0 {
			t = time.Duration(0)
		} else if src[i-1].time < sendDuration*2/10 {
			t = src[i-1].time + time.Duration(baseTimeDelta*3/2)
		} else if src[i-1].time < sendDuration*8/10 {
			t = src[i-1].time + time.Duration(baseTimeDelta*3/4)
		} else if src[i-1].time < sendDuration {
			t = src[i-1].time + time.Duration(baseTimeDelta*3/2)
		} else {
			break
		}

		//if i <= allTasksNum/4 {
		//	t = time.Duration(int64(i)*10*int64(time.Second)/int64(allTasksNum)) * 3 / 2
		//} else if i <= allTasksNum*3/4 {
		//	t = src[allTasksNum/4].time
		//	t += time.Duration(int64(i-(allTasksNum/4)) * 10 * int64(time.Second) / int64(allTasksNum) * 3 / 4)
		//} else {
		//	t = src[allTasksNum*3/4].time
		//	t += time.Duration(int64(i-(allTasksNum*3/4)) * 10 * int64(time.Second) / int64(allTasksNum) * 3 / 2)
		//}
		src = append(src, SrcNode{time.Duration(t), newTask})
	}
	return src
}

// trace file

// Component For NodeComponent interface
func (n TaskGen) Component() ComponentName { return CTaskGen }

// SetOsApi for NodeComponent interface
func (n *TaskGen) SetOsApi(osapi OsApi) { n.Os = osapi }

func (n *TaskGen) Debug() {}

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

func (s *Scheduler) Debug() {
	log.Println(s.WaitSchedule)
	log.Println(s.TasksStatus)
}

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

func (s *StateStorage) Debug() { log.Println(s.Workers) }

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

func (n *ResourceManager) Debug() { log.Println(n.Tasks) }

type RaftRole string

const (
	Follower  RaftRole = "Follower"
	Leader    RaftRole = "Leader"
	Candidate RaftRole = "Candidate"
)

// RaftManager 组件
type RaftManager struct {
	Os            OsApi
	Host          string
	IsBroken      bool
	AllNodeNum    int
	Role          RaftRole
	StartTime     time.Time
	LeaderTime    time.Time
	LastHeartBeat time.Time     // the last time of leader heartbeat
	LeaderTimeout time.Duration // how long for judging  the leader is dead
	LeaderAddr    string
	Term          int
	ReceiveYES    int
	ReceiveNO     int
}

// NewRaftManager 创建新的Raft节点组件
func NewRaftManager(host string) *RaftManager {
	return &RaftManager{
		Host: host,
	}
}

// Component For NodeComponent interface
func (r RaftManager) Component() ComponentName { return CRaftManager }

// SetOsApi For NodeComponent interface
func (r *RaftManager) SetOsApi(osapi OsApi) { r.Os = osapi }

func (r *RaftManager) Debug() {}
