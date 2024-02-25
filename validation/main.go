package main

import (
	"fmt"
	"log"
	"simds-standalone/config"
	"simds-standalone/core"
	"time"
)

var tasknumPerSec int32

type scheduler struct {
	input chan core.TaskInfo
}

func NewScheduler(tasknum int32) scheduler {
	return scheduler{input: make(chan core.TaskInfo, tasknum)}
}

func (s *scheduler) run(logchan chan string) {
	var scheduler core.Scheduler
	scheduler.Workers = make(map[string]*core.NodeInfo)
	workernum := config.Val.NodeNum
	for i := 0; i < int(workernum); i++ {
		nodeAddr := "worker" + fmt.Sprint(i) + ":" + string(core.CResouceManger)
		nodeinfo := &core.NodeInfo{nodeAddr, config.Val.NodeCpu, config.Val.NodeMemory, 10, 10}
		scheduler.Workers["worker"+fmt.Sprint(i)+":"+string(core.CResouceManger)] = nodeinfo.Clone()
	}

	for i := int64(0); ; i++ {
		task := <-s.input

		workerid := ""
		workerLeftCpu := int32(0)
		for _, workerinfo := range scheduler.Workers {
			if task.CpuRequest < workerinfo.Cpu && task.MemoryRequest < workerinfo.Memory {
				if workerinfo.Cpu > workerLeftCpu {
					workerid = workerinfo.Addr
					workerLeftCpu = workerinfo.Cpu
				}
			}
		}

		// send task allocate

		logchan <- fmt.Sprintf("%v,%s,%s,%s,%d,%d", time.Now().Format(time.RFC3339Nano), task.Id, "start", workerid, task.CpuRequest, task.MemoryRequest)
	}
}

func taskgen(rec chan core.TaskInfo, logchan chan string, tasknum int32) {
	start := time.Now()
	for i := int32(0); i < tasknum; i++ {
		oneSendstart := time.Now()
		task := core.TaskInfo{Id: fmt.Sprintf("task%d", i), CpuRequest: 300, MemoryRequest: 300}
		rec <- task
		logchan <- fmt.Sprintf("%v,%s,%s,%s,%d,%d", time.Now().Format(time.RFC3339Nano), task.Id, "submit", "taskgen", task.CpuRequest, task.MemoryRequest)

		if i%(tasknumPerSec/1000) == 0 {
			for time.Since(oneSendstart) < time.Millisecond {

			}
		}
	}

	log.Println("taskgen finished", time.Since(start))

}

func main() {
	tasknumPerSec = int32(float32(config.Val.NodeNum) * config.Val.TaskNumFactor)
	tasknum := tasknumPerSec * config.Val.SimulateDuration / 1000
	log.Println("tasknumPerSec",tasknumPerSec)
	s := NewScheduler(tasknum)

	log := make(chan string, tasknum*2)

	go s.run(log)
	taskgen(s.input, log, tasknum)

	fmt.Println("time,taskid,type,nodeip,cpu,ram")
	for i := int32(0); i < 2*tasknum; i++ {
		l := <-log
		fmt.Println(l)
	}

}
