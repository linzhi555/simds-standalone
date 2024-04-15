package core

import (
	"fmt"
	"simds-standalone/config"
	"time"
)

type Worker struct {
	Os       OsApi
	HostName string
	Tasks    map[string]*TaskInfo
	Node     NodeInfo // do not store the information , calculate when needed from tasks
	Manager  string   // will nofify the ther worker's Manager if the task state is changed
}

func NewWorker(host string, nodeinfo NodeInfo, manager string) *Worker {
	return &Worker{
		HostName: host,
		Manager:  manager,
		Tasks:    make(map[string]*TaskInfo),
	}
}

func (n *Worker) SetOsApi(osapi OsApi) { n.Os = osapi }
func (n *Worker) Debug() {
	res := fmt.Sprintf("%+v\n", n.Node)
	res += fmt.Sprintln("tasks:")
	for _, task := range n.Tasks {
		res += fmt.Sprintf("%+v\n", *task)
	}
	fmt.Println(res)
}

func (worker *Worker) Setup() {

}

// worker 运行逻辑
// 1. 直接运行,包括分配和运行
// 2. 先分配不运行，
// 3. 运行之前预分配，
// 4. 取消之前预分配，
// 同时监控任务状态，在任务结束后发送任务结束信息给TaskFinishReceiver
// 该TaskFinishReceiver 可在组件初始化函数中指定
func (worker *Worker) Update() {

	hostTime := worker.Os.GetTime()
	if !worker.Os.Net().Empty() {
		newMessage, err := worker.Os.Net().Recv()
		if err != nil {
			panic(err)
		}
		worker.Os.LogInfo(worker.Os, "received message:", newMessage)

		if newMessage.Content == "TaskStart" {
			taskid := newMessage.Body.(TaskInfo).Id
			if t, ok := worker.Tasks[taskid]; ok {
				if t.Status == "needStart" {
					t.Status = "start"
					t.StartTime = hostTime
					t.LeftTime = t.LifeTime
					worker.Os.LogInfo(worker.Os, "start task:", t)
					worker.Os.LogInfo(hostTime, t, worker.Os.Net().GetAddr())
				}
			}
		}

		if newMessage.Content == "TaskPreAllocate" {
			newTask := newMessage.Body.(TaskInfo)
			worker.Tasks[newTask.Id] = &newTask
			newTask.Status = "needStart"
			worker.Os.LogInfo(worker.Os, "allocate task:", newTask)
		}

		if newMessage.Content == "TaskRun" {
			newTask := newMessage.Body.(TaskInfo)
			newTask.StartTime = hostTime
			newTask.LeftTime = newTask.LifeTime
			worker.Tasks[newTask.Id] = &newTask
			newTask.Status = "start"
			worker.Os.LogInfo(worker.Os, "Start task:", newTask)
			worker.Os.LogInfo(hostTime, &newTask, worker.Os.Net().GetAddr())
		}

		if newMessage.Content == "TaskCancelAllocate" {
			taskid := newMessage.Body.(TaskInfo).Id
			if t, ok := worker.Tasks[taskid]; ok {
				if t.Status == "needStart" {
					t.Status = "finish"
					worker.Os.LogInfo(worker.Os, "cancel task:", t)

					if worker.Manager != "" {
						informReceiverTaskStatus(worker, t, "TaskFinish")
					}
					delete(worker.Tasks, taskid)

				}
			}
		}

	}
	for id, t := range worker.Tasks {
		if t.Status == "start" {
			if t.LeftTime > 0 {
				t.LeftTime -= time.Second / time.Duration(config.Val.FPS)

			} else {
				t.Status = "finish"
				worker.Os.LogInfo(hostTime, t, worker.Os.Net().GetAddr())
				if worker.Manager != "" {
					informReceiverTaskStatus(worker, t, "TaskFinish")
				}
				delete(worker.Tasks, id)
				nodeinfo := _calculateNodeInfo(worker)
				worker.Os.LogInfo(worker.Os, "Task Finished", t, "now, nodeinfo is", nodeinfo)
			}
		}
	}
	worker.Node = _calculateNodeInfo(worker)
}

func informReceiverTaskStatus(worker *Worker, t *TaskInfo, content string) {
	newMessage := Message{
		From:    worker.Os.Net().GetAddr(),
		To:      worker.Manager,
		Content: content,
		Body:    *t,
	}
	err := worker.Os.Net().Send(newMessage)
	if err != nil {
		panic(err)
	}

}

func _calculateNodeInfo(worker *Worker) NodeInfo {
	var cpu int32 = 0
	var memory int32 = 0

	for _, t := range worker.Tasks {
		cpu += t.CpuRequest
		memory += t.MemoryRequest
	}

	var nodeinfo NodeInfo = NodeInfo{
		Addr:           worker.Os.Net().GetAddr(),
		Cpu:            config.Val.NodeCpu,
		Memory:         config.Val.NodeMemory,
		CpuAllocted:    cpu,
		MemoryAllocted: memory,
	}
	return nodeinfo
}
