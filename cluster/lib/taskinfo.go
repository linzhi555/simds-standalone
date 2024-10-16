package lib

import (
	"encoding/json"
	"log"
	"time"

	"simds-standalone/cluster/base"
)

func init() {
	base.InverseJsonTable["Task*"] = func(s string) base.MessageBody {
		var res TaskInfo
		err := json.Unmarshal([]byte(s), &res)
		if err != nil {
			log.Panicln("try to unmarshal TaskInfo:", s, err)
		}
		return res
	}
}

// TaskInfo 任务的基本信息，还有一些附加的调度器使用的字段
type TaskInfo struct {
	Id                string //the task id,it is unique
	CpuRequest        int32
	MemoryRequest     int32
	StartTime         time.Time
	LifeTime          time.Duration
	LeftTime          time.Duration
	Status            string
	User              string
	Worker            string
	ScheduleFailCount int32
	Cmd               string
}

// MessageBody TaskInfo 是MessageBody
func (TaskInfo) MessageBody() {}

// Clone 复制新的TaskInfo
func (t *TaskInfo) Clone() *TaskInfo {
	var newT TaskInfo = *t
	return &newT
}
