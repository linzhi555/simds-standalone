package main

import (
	"fmt"
	"os"
	"path"
	"simds-standalone/common"
	"time"
)

func initTaskEventLog() {
	f, err := os.OpenFile(path.Join(Config.OutputDir, "./tasks_event.log"), os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	_, err = f.WriteString("time,taskid,type,nodeip,cpu,ram\n")
	if err != nil {
		panic(err)
	}
}

// TaskEventLog 任务转台信息输出至csv
// 格式文件，由于任务延迟和集群状态分析
func TaskEventLog(t time.Time, task *TaskInfo, host string) {
	timestr := t.Format(time.RFC3339Nano)
	err := common.AppendLineCsvFile(path.Join(Config.OutputDir, "./tasks_event.log"), []string{timestr, task.Id, task.Status, string(host), fmt.Sprint(task.CpuRequest), fmt.Sprint(task.MemoryRequest)})
	if err != nil {
		panic(err)
	}
}
