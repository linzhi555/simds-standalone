package main

import (
	"fmt"
	"os"
	"simds-standalone/common"
	"time"
)

func BuildCenterCluster() Cluster {
	var cluster = createCluster()
	var nodes []Node
	nodes = append(nodes, Node{
		"user1",
		[]NodeComponent{
			NewTaskGen("user1"),
		},
	})

	nodes = append(nodes, Node{
		"master1",
		[]NodeComponent{
			NewScheduler("master1"),
		},
	})
	for i := 0; i < int(Config.NodeNum); i++ {

		workerName := fmt.Sprintf("worker%d", i)
		nodes = append(nodes, Node{
			workerName,
			[]NodeComponent{
				NewResourceManager(workerName),
			},
		})

	}
	cluster.Nodes = nodes
	cluster.RegisterFunc(CTaskGen, CenterTaskgen_setup, CommonTaskgenUpdate)
	cluster.RegisterFunc(CScheduler, CenterSchedulerSetup, CenterSchedulerUpdate)
	cluster.RegisterFunc(CResouceManger, CenterResourceManagerSetup, CommonResourceManagerUpdate)

	return cluster
}

func CenterTaskgen_setup(c interface{}) {
	taskgen := c.(*TaskGen)
	taskgen.StartTime = taskgen.Os.GetTime()
	taskgen.Receivers = append(taskgen.Receivers, "master1"+":"+string(CScheduler))
}

func CenterSchedulerSetup(comp interface{}) {
	scheduler := comp.(*Scheduler)
	for i := 0; i < int(Config.NodeNum); i++ {
		nodeinfo := &NodeInfo{Config.NodeCpu, Config.NodeMemory, 0, 0}
		scheduler.Workers["worker"+fmt.Sprint(i)+":"+string(CResouceManger)] = nodeinfo.Clone()
	}

}

func CenterSchedulerUpdate(comp interface{}) {

	scheduler := comp.(*Scheduler)

	for !scheduler.Os.Net().Empty() {
		newMessage, err := scheduler.Os.Net().Recv()
		if err != nil {
			panic(err)
		}

		if newMessage.Content == "TaskDispense" {
			task := newMessage.Body.(TaskInfo)
			task.Status = "WaitSchedule"
			scheduler.WaitSchedule.InQueue(task)
			LogInfo(scheduler.Os, "received TaskDispense", task)
		}

		if newMessage.Content == "TaskFinish" {
			taskInfo := newMessage.Body.(TaskInfo)
			scheduler.Workers[newMessage.From].SubAllocated(taskInfo.CpuRequest, taskInfo.MemoryRequest)
			LogInfo(scheduler.Os, "received TaskFinish", newMessage.From, taskInfo)
		}

	}

	var MAX_SCHEDULE_TIMES = int(Config.SchedulerPerformance)
	for i := 0; i < MAX_SCHEDULE_TIMES; i++ {

		task, err := scheduler.WaitSchedule.Dequeue()
		if err != nil {
			break
		}

		dstWorker, ok := schdulingAlgorithm(scheduler, &task)
		if ok {
			task.Worker = dstWorker
			task.Status = "Allocated"
			scheduler.Workers[task.Worker].AddAllocated(task.CpuRequest, task.MemoryRequest)
			newMessage := Message{
				From:    scheduler.Os.Net().GetAddr(),
				To:      task.Worker,
				Content: "TaskRun",
				Body:    task,
			}
			scheduler.Os.Net().Send(newMessage)
			LogInfo(scheduler.Os, "sendtask to", task.Worker, task)
		} else {
			scheduler.WaitSchedule.InQueueFront(task)

		}

	}
}

// schedule the task,if it  can not find a worker for the task,return "",false
// else return "addr of some worker",true
func schdulingAlgorithm(scheduler *Scheduler, task *TaskInfo) (dstAddr string, ok bool) {
	dstAddr = ""

	keys := make([]string, 0, len(scheduler.Workers))
	for k := range scheduler.Workers {
		keys = append(keys, k)
	}
	common.ShuffleStringSlice(keys)
	for _, workerAddr := range keys {
		nodeinfo := scheduler.Workers[workerAddr]
		if nodeinfo.CanAllocate(task.CpuRequest, task.MemoryRequest) {
			dstAddr = workerAddr
		}
	}

	if dstAddr == "" {
		return dstAddr, false
	}
	return dstAddr, true
}

func CenterResourceManagerSetup(comp interface{}) {
	rm := comp.(*ResourceManager)
	rm.TaskFinishReceiver = "master1" + ":" + string(CScheduler)

}

func init() {
	f, err := os.OpenFile("./test.log", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	f.WriteString("time,taskid,type,nodeip,cpu,ram\n")
	f.Close()
}

func TaskEventLog(t time.Time, task *TaskInfo, host string) {
	timestr := t.Format(time.RFC3339Nano)
	common.AppendLineCsvFile("./test.log", []string{timestr, task.Id, task.Status, string(host), fmt.Sprint(task.CpuRequest), fmt.Sprint(task.MemoryRequest)})
}
