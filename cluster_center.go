package main

import (
	"fmt"
	"math/rand"
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
			CreateTaskGen("user1"),
		},
	})

	nodes = append(nodes, Node{
		"master1",
		[]NodeComponent{
			CreateScheduler("master1"),
		},
	})
	for i := 0; i < int(Config.NodeNum); i++ {

		workerName := fmt.Sprintf("worker%d", i)
		nodes = append(nodes, Node{
			workerName,
			[]NodeComponent{
				CreateResourceManager(workerName),
			},
		})

	}
	cluster.Nodes = nodes
	cluster.RegisterFunc(CTaskGen, taskgen_setup, taskgen_update)
	cluster.RegisterFunc(CScheduler, scheduler_setup, scheduler_update)
	cluster.RegisterFunc(CResouceManger, resourceManager_setup, resourceManager_update)

	return cluster
}

func taskgen_setup(c interface{}) {
	taskgen := c.(*TaskGen)
	taskgen.StartTime = taskgen.GetTime()
}

func taskgen_update(c interface{}) {
	taskgen := c.(*TaskGen)
	t := taskgen.GetTime().Sub(taskgen.StartTime)

	taskNumPerSecond := Config.TaskNumFactor * float32(Config.NodeNum)

	taskgenAddr := "user1" + ":" + string(CTaskGen)
	masterAddr := "master1" + ":" + string(CScheduler)

	if t < 10*time.Second {
		for taskgen.CurTaskId < int(taskNumPerSecond*float32(t.Milliseconds())/float32(Second)) {

			newtask := TaskInfo{
				Id:            fmt.Sprintf("task%d", taskgen.CurTaskId),
				CpuRequest:    1 + int32(rand.Intn(int(2*Config.TaskCpu-2))),
				MemoryRequest: 1 + int32(rand.Intn(int(2*Config.TaskMemory-2))),
				LifeTime:      time.Duration(100+int32(rand.Intn(int(Config.TaskLifeTime)*2-200))) * time.Millisecond,
				Status:        "submit",
			}

			newMessage := Message{
				From:    taskgenAddr,
				To:      masterAddr,
				Content: "TaskDispense",
				Body:    newtask,
			}
			taskgen.Net().Send(newMessage)
			LogInfo(taskgen, fmt.Sprintf(": send task to %s %v", masterAddr, newMessage.Body))
			TaskEventLog(taskgen.GetTime(), &newtask, masterAddr)
			taskgen.CurTaskId += 1
		}
	}
}

func scheduler_setup(comp interface{}) {
	scheduler := comp.(*Scheduler)
	for i := 0; i < int(Config.NodeNum); i++ {
		nodeinfo := &NodeInfo{Config.NodeCpu, Config.NodeMemory, 0, 0}
		scheduler.Workers["worker"+fmt.Sprint(i)+":"+string(CResouceManger)] = nodeinfo.Clone()
	}

}

func scheduler_update(comp interface{}) {

	scheduler := comp.(*Scheduler)

	for !scheduler.Net().Empty() {
		newMessage, err := scheduler.net.Recv()
		if err != nil {
			panic(err)
		}

		if newMessage.Content == "TaskDispense" {
			task := newMessage.Body.(TaskInfo)
			task.Status = "WaitSchedule"
			scheduler.WaitSchedule.InQueue(task)
			LogInfo(scheduler, "received TaskDispense", task)
		}

		if newMessage.Content == "TaskFinish" {
			taskInfo := newMessage.Body.(TaskInfo)
			scheduler.Workers[newMessage.From].SubAllocated(taskInfo.CpuRequest, taskInfo.MemoryRequest)
			LogInfo(scheduler, "received TaskFinish", newMessage.From, taskInfo)
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
				From:    scheduler.Net().GetAddr(),
				To:      task.Worker,
				Content: "TaskAllocate",
				Body:    task,
			}
			scheduler.Net().Send(newMessage)
			LogInfo(scheduler, "sendtask to", task.Worker, task)
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

func resourceManager_setup(comp interface{}) {
	rm := comp.(*ResourceManager)
	rm.TaskFinishReceiver = "master1" + ":" + string(CScheduler)

}

func resourceManager_update(comp interface{}) {

	rm := comp.(*ResourceManager)
	hostTime := rm.GetTime()

	if !rm.Net().Empty() {
		newMessage, err := rm.Net().Recv()
		if err != nil {
			panic(err)
		}
		LogInfo(rm, "received message:", newMessage)

		if newMessage.Content == "TaskAllocate" {
			newTask := newMessage.Body.(TaskInfo)
			newTask.StartTime = hostTime
			rm.Tasks[newTask.Id] = &newTask
			newTask.Status = "start"
			LogInfo(rm, "Start task:", newTask)
			TaskEventLog(hostTime, &newTask, rm.Net().GetAddr())
		}
	}

	for id, t := range rm.Tasks {
		if t.Status == "start" && hostTime.After(t.StartTime.Add(t.LifeTime)) {
			t.Status = "finish"
			LogInfo(rm, "Task Finished", t)
			TaskEventLog(hostTime, t, rm.net.GetAddr())
			if rm.TaskFinishReceiver != "" {
				informReceiverTaskStatus(rm, t, "TaskFinish")
			}

			delete(rm.Tasks, id)
		}
	}

	updateNodeInfo(rm)
}

func informReceiverTaskStatus(rm *ResourceManager, t *TaskInfo, content string) {
	newMessage := Message{
		From:    rm.Net().GetAddr(),
		To:      rm.TaskFinishReceiver,
		Content: content,
		Body:    *t,
	}
	rm.Net().Send(newMessage)
}

func updateNodeInfo(rm *ResourceManager) {
	var cpu int32 = 0
	var memory int32 = 0

	for _, t := range rm.Tasks {
		cpu += t.CpuRequest
		memory += t.MemoryRequest
	}

	if rm.Node.CpuAllocted != cpu || rm.Node.MemoryAllocted != memory {
		rm.Node.CpuAllocted = cpu
		rm.Node.MemoryAllocted = memory
	}
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
