package main

import (
	"fmt"
	"os"
	"simds-standalone/common"
	"time"
)

// CommonTaskgenUpdate 通用的任务发生器，适
// 用多种形式集群，通过Receivers字段来确定发送对象
func CommonTaskgenUpdate(c interface{}) {
	taskgen := c.(*TaskGen)
	timeNow := taskgen.Os.GetTime().Sub(taskgen.StartTime)
	if timeNow <= 0 {
		return
	}

	taskgenAddr := taskgen.Os.Net().GetAddr()

	receiverNum := len(taskgen.Receivers)

	for taskgen.CurTaskId < len(taskgen.Src) {
		if taskgen.Src[taskgen.CurTaskId].time > timeNow {
			break
		}

		newtask := taskgen.Src[taskgen.CurTaskId].task
		receiverAddr := taskgen.Receivers[taskgen.CurTaskId%receiverNum]
		newMessage := Message{
			From:    taskgenAddr,
			To:      receiverAddr,
			Content: "TaskDispense",
			Body:    newtask,
		}
		err := taskgen.Os.Net().Send(newMessage)
		if err != nil {
			panic(err)
		}

		LogInfo(taskgen.Os, fmt.Sprintf(": send task to %s %v", receiverAddr, newMessage.Body))
		TaskEventLog(taskgen.Os.GetTime(), &newtask, receiverAddr)
		taskgen.CurTaskId++
	}
}

// CommonResourceManagerUpdate 通用的任务资源管理器
// ，适用多种形式集群，接收下列信息-
// 1. 直接运行,包括分配和运行
// 2. 先分配不运行，
// 3. 运行之前预分配，
// 4. 取消之前预分配，
// 同时监控任务状态，在任务结束后发送任务结束信息给TaskFinishReceiver
// 该TaskFinishReceiver 可在组件初始化函数中指定
func CommonResourceManagerUpdate(comp interface{}) {

	rm := comp.(*ResourceManager)
	hostTime := rm.Os.GetTime()

	if !rm.Os.Net().Empty() {
		newMessage, err := rm.Os.Net().Recv()
		if err != nil {
			panic(err)
		}
		LogInfo(rm.Os, "received message:", newMessage)

		if newMessage.Content == "TaskStart" {
			taskid := newMessage.Body.(TaskInfo).Id
			if t, ok := rm.Tasks[taskid]; ok {
				if t.Status == "needStart" {
					t.Status = "start"
					t.StartTime = hostTime
					LogInfo(rm.Os, "start task:", t)
					TaskEventLog(hostTime, t, rm.Os.Net().GetAddr())
				}
			}
		}

		if newMessage.Content == "TaskPreAllocate" {
			newTask := newMessage.Body.(TaskInfo)
			rm.Tasks[newTask.Id] = &newTask
			newTask.Status = "needStart"
			LogInfo(rm.Os, "allocate task:", newTask)
		}

		if newMessage.Content == "TaskRun" {
			newTask := newMessage.Body.(TaskInfo)
			newTask.StartTime = hostTime
			rm.Tasks[newTask.Id] = &newTask
			newTask.Status = "start"
			LogInfo(rm.Os, "Start task:", newTask)
			TaskEventLog(hostTime, &newTask, rm.Os.Net().GetAddr())
		}

		if newMessage.Content == "TaskCancelAllocate" {
			taskid := newMessage.Body.(TaskInfo).Id
			if t, ok := rm.Tasks[taskid]; ok {
				if t.Status == "needStart" {
					t.Status = "finish"
					LogInfo(rm.Os, "cancel task:", t)

					if rm.TaskFinishReceiver != "" {
						informReceiverTaskStatus(rm, t, "TaskFinish")
					}
					delete(rm.Tasks, taskid)

				}
			}
		}

	}

	for id, t := range rm.Tasks {
		if t.Status == "start" && hostTime.After(t.StartTime.Add(t.LifeTime)) {
			t.Status = "finish"
			TaskEventLog(hostTime, t, rm.Os.Net().GetAddr())
			if rm.TaskFinishReceiver != "" {
				informReceiverTaskStatus(rm, t, "TaskFinish")
			}
			delete(rm.Tasks, id)
			nodeinfo := _calculateNodeInfo(rm)
			LogInfo(rm.Os, "Task Finished", t, "now, nodeinfo is", nodeinfo)
		}
	}

}

func informReceiverTaskStatus(rm *ResourceManager, t *TaskInfo, content string) {
	newMessage := Message{
		From:    rm.Os.Net().GetAddr(),
		To:      rm.TaskFinishReceiver,
		Content: content,
		Body:    *t,
	}
	err := rm.Os.Net().Send(newMessage)
	if err != nil {
		panic(err)
	}

}

func _calculateNodeInfo(rm *ResourceManager) NodeInfo {
	var cpu int32 = 0
	var memory int32 = 0

	for _, t := range rm.Tasks {
		cpu += t.CpuRequest
		memory += t.MemoryRequest
	}

	var nodeinfo NodeInfo = NodeInfo{
		Addr:           rm.Os.Net().GetAddr(),
		Cpu:            Config.NodeCpu,
		Memory:         Config.NodeMemory,
		CpuAllocted:    cpu,
		MemoryAllocted: memory,
	}
	return nodeinfo
}

func init() {
	f, err := os.OpenFile("./tasks_event.log", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
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
	err := common.AppendLineCsvFile("./tasks_event.log", []string{timestr, task.Id, task.Status, string(host), fmt.Sprint(task.CpuRequest), fmt.Sprint(task.MemoryRequest)})
	if err != nil {
		panic(err)
	}
}
