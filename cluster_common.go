package main

import (
	"fmt"
	"math/rand"
	"time"
)

func CommonTaskgen_update(c interface{}) {
	taskgen := c.(*TaskGen)
	t := taskgen.Os.GetTime().Sub(taskgen.StartTime)

	taskNumPerSecond := Config.TaskNumFactor * float32(Config.NodeNum)

	taskgenAddr := "user1" + ":" + string(CTaskGen)

	receiverNum := len(taskgen.Receivers)

	if t < 10*time.Second {
		for taskgen.CurTaskId < int(taskNumPerSecond*float32(t.Milliseconds())/float32(Second)) {

			newtask := TaskInfo{
				Id:            fmt.Sprintf("task%d", taskgen.CurTaskId),
				CpuRequest:    1 + int32(rand.Intn(int(2*Config.TaskCpu-2))),
				MemoryRequest: 1 + int32(rand.Intn(int(2*Config.TaskMemory-2))),
				LifeTime:      time.Duration(100+int32(rand.Intn(int(Config.TaskLifeTime)*2-200))) * time.Millisecond,
				Status:        "submit",
			}

			receiverAddr := taskgen.Receivers[taskgen.CurTaskId%receiverNum]
			newMessage := Message{
				From:    taskgenAddr,
				To:      receiverAddr,
				Content: "TaskDispense",
				Body:    newtask,
			}
			taskgen.Os.Net().Send(newMessage)
			LogInfo(taskgen.Os, fmt.Sprintf(": send task to %s %v", receiverAddr, newMessage.Body))
			TaskEventLog(taskgen.Os.GetTime(), &newtask, receiverAddr)
			taskgen.CurTaskId += 1
		}
	}
}

func CommonResourceManager_update(comp interface{}) {

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
			LogInfo(rm.Os, "Task Finished", t)
			TaskEventLog(hostTime, t, rm.Os.Net().GetAddr())
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
		From:    rm.Os.Net().GetAddr(),
		To:      rm.TaskFinishReceiver,
		Content: content,
		Body:    *t,
	}
	rm.Os.Net().Send(newMessage)
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
