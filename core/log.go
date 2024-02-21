package core

import (
	"fmt"
	"os"
	"path"
	"simds-standalone/common"
	"simds-standalone/config"
	"time"
)

const TASKS_EVENT_LOG_NAME = "tasks_event.log"
const NETWORK_EVENT_LOG_NAME = "network_event.log"

func InitLogs() {
	// init tasks_event.log
	f, err := os.OpenFile(path.Join(config.Val.OutputDir, TASKS_EVENT_LOG_NAME), os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	_, err = f.WriteString("time,taskid,type,nodeip,cpu,ram\n")
	if err != nil {
		panic(err)
	}
	f.Close()

	// init networks.log
	f, err = os.OpenFile(path.Join(config.Val.OutputDir, NETWORK_EVENT_LOG_NAME), os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	_, err = f.WriteString("time_ms,type,from,to\n")
	if err != nil {
		panic(err)
	}
	f.Close()
}

// TaskEventLog 任务转态信息输出至csv
// 格式文件，由于任务延迟和集群状态分析
func TaskEventLog(t time.Time, task *TaskInfo, host string) {
	timestr := t.Format(time.RFC3339Nano)
	err := common.AppendLineCsvFile(path.Join(config.Val.OutputDir, TASKS_EVENT_LOG_NAME), []string{timestr, task.Id, task.Status, string(host), fmt.Sprint(task.CpuRequest), fmt.Sprint(task.MemoryRequest)})
	if err != nil {
		panic(err)
	}
}

// 记录一次网络通信
func NetEventLog(t int64, eventType string, message *Message) {
	timestr := fmt.Sprint(t)
	err := common.AppendLineCsvFile(path.Join(config.Val.OutputDir, NETWORK_EVENT_LOG_NAME), []string{timestr, eventType, message.From, message.To})
	if err != nil {
		panic(err)
	}
}
