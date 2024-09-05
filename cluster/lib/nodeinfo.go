package lib

import (
	"encoding/json"
	"simds-standalone/cluster/base"
)

func init() {
	base.InverseJsonTable["NodeInfo*"] = func(s string) base.MessageBody {
		var res NodeInfo
		err := json.Unmarshal([]byte(s), &res)
		if err != nil {
			panic(err)
		}
		return res
	}

	base.InverseJsonTable["VecNodeInfo*"] = func(s string) base.MessageBody {
		var res base.Vec[NodeInfo]
		err := json.Unmarshal([]byte(s), &res)
		if err != nil {
			panic(err)
		}
		return res
	}

}

// NodeInfo  节点资源信息

type NodeInfo struct {
	Addr           string
	Cpu            int32
	Memory         int32
	CpuAllocted    int32
	MemoryAllocted int32
}

// MessageBody NodeInfo 是MessageBody
func (NodeInfo) MessageBody() {}

// Clone 复制新的NodeInfo
func (n *NodeInfo) Clone() *NodeInfo {
	var NodeInfoCopy = *n
	return &NodeInfoCopy
}

func (n *NodeInfo) CpuPercent() float32 {
	return float32(n.CpuAllocted) / float32(n.Cpu)
}

func (n *NodeInfo) MemoryPercent() float32 {
	return float32(n.MemoryAllocted) / float32(n.Memory)
}

// AddAllocated 更新节点信息-增加已分配
func (n *NodeInfo) AddAllocated(taskCpu, taskMemory int32) {
	n.CpuAllocted += taskCpu
	n.MemoryAllocted += taskMemory
}

// SubAllocated 更新节点信息-释放已分配
func (n *NodeInfo) SubAllocated(taskCpu, taskMemory int32) {
	n.CpuAllocted -= taskCpu
	n.MemoryAllocted -= taskMemory
}

// CanAllocate 判读是否满足分配
func (n *NodeInfo) CanAllocate(taskCpu, taskMemory int32) bool {
	if n.Cpu-n.CpuAllocted >= taskCpu && n.Memory-n.MemoryAllocted >= taskMemory {
		return true
	}
	return false
}

// CanAllocateTask 判读是否满足分配某个任务
func (n *NodeInfo) CanAllocateTask(task *TaskInfo) bool {
	return n.CanAllocate(task.CpuRequest, task.MemoryRequest)
}
