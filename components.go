package main

type SystemTime int32
func (s *SystemTime) Component() string {
	return "SystemTime"
}

type TaskInfo struct {
	CpuRequest    int32
	MemoryRequest int32
	StartTime     int32
	LifeTime      int32
	Status        string
}
func (t * TaskInfo)Component()string{
	return "TaskInfo"

}
type NodeInfo struct {
	Cpu            int32
	Memory         int32
	CpuAllocted    int32
	MemoryAllocted int32
}

func (n *NodeInfo) Component() string {
	return "Node"
}

type TaskList struct {
	AllTasks []*TaskInfo
}

func (n *TaskList) Component() string {
	return "TaskList"
}


