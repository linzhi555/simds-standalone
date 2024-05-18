package core

import (
	"fmt"
	"os/exec"
	"simds-standalone/config"
	"time"
)

type Worker struct {
	BasicNode
	Node    NodeInfo // do not store the information , calculate when needed from tasks
	Manager string   // will nofify the ther worker's Manager if the task state is changed
	TaskMap map[string]*TaskInfo
}

func NewWorker(host string, nodeinfo NodeInfo, manager string) *Worker {
	return &Worker{
		BasicNode: BasicNode{Host: host},
		Manager:   manager,
		Node:      nodeinfo,
		TaskMap:   map[string]*TaskInfo{},
	}
}

func (n *Worker) Debug() {
	// res := fmt.Sprintf("%+v\n", n.Node)
	// res += fmt.Sprintln("tasks:")
	// for _, task := range n.Tasks {
	// 	res += fmt.Sprintf("%+v\n", n.TaskManager.Tasks)
	// }
	// fmt.Println(res)
}

// worker 运行逻辑
// 收到其他节点信息处理：
// 1. 直接运行命令
// 2. 先分配不运行命令
// 3. 运行之前预分配命令，
// 4. 取消之前预分配命令，
// 收到后台任务管理器的信息
// 1. 任务介绍信息，则将此消息通知给Manager

func (worker *Worker) Update(msg Message) {

	switch msg.Content {

	case "TaskStart":

		taskid := msg.Body.(TaskInfo).Id
		if t, ok := worker.TaskMap[taskid]; ok {
			if t.Status == "needStart" {
				t.Status = "start"
				t.StartTime = worker.Os.GetTime()
				t.LeftTime = t.LifeTime

				worker.Os.LogInfo(TASKS_EVENT_LOG_NAME, t.Id, "start", worker.GetHostName(), fmt.Sprint(t.CpuRequest), fmt.Sprint(t.MemoryRequest))
				worker.Os.Send(Message{
					From:    worker.Host,
					To:      t.User,
					Content: "TaskStart",
					Body:    t,
				})
			}
		}
	case "TaskPreAllocate":
		newTask := msg.Body.(TaskInfo)
		worker.TaskMap[newTask.Id] = &newTask
		newTask.Status = "needStart"
	case "TaskRun":
		newTask := msg.Body.(TaskInfo)
		newTask.StartTime = worker.Os.GetTime()
		newTask.LeftTime = newTask.LifeTime
		worker.TaskMap[newTask.Id] = &newTask
		newTask.Status = "start"
		worker._runTask(newTask)
		worker.Os.Send(Message{
			From:    worker.Host,
			To:      newTask.User,
			Content: "TaskStart",
			Body:    newTask,
		})

	case "TaskCancelAlloc":
		taskid := msg.Body.(TaskInfo).Id
		if t, ok := worker.TaskMap[taskid]; ok {
			if t.Status == "needStart" {
				t.Status = "finish"
				if worker.Manager != "" {
					informReceiverTaskStatus(worker, t, "TaskFinish")
				}
				delete(worker.TaskMap, taskid)

			}
		}
	case "TaskFinish":
		t := msg.Body.(TaskInfo)
		id := t.Id
		if worker.Manager != "" {
			informReceiverTaskStatus(worker, &t, "TaskFinish")
		}
		delete(worker.TaskMap, id)
		nodeinfo := _calculateNodeInfo(worker)
		worker.Os.LogInfo("stdout", worker.GetHostName(), "TaskFinish", fmt.Sprint(nodeinfo))
		worker.Os.Send(Message{
			From:    worker.Host,
			To:      t.User,
			Content: "TaskFinish",
			Body:    t,
		})
	}

}

func (node *Worker) _runTask(t TaskInfo) {
	node.Os.Run(func() {
		cmd := exec.Command("bash", "-c", t.Cmd)
		err := cmd.Run()
		if err != nil {
			panic(err)
		}
		newMessage := Message{
			From:    node.GetHostName(),
			To:      node.GetHostName(),
			Content: "TaskFinish",
			Body:    t,
		}
		err = node.Os.Send(newMessage)
		if err != nil {
			panic(err)
		}

	})
}

func (worker *Worker) SimulateTasksUpdate() {
	for _, t := range worker.TaskMap {
		if t.Status == "start" && t.LeftTime > 0 {
			t.LeftTime -= (time.Second / time.Duration(config.Val.FPS))
			if t.LeftTime <= 0 {
				newMessage := Message{
					From:    worker.GetHostName(),
					To:      worker.GetHostName(),
					Content: "TaskFinish",
					Body:    *t,
				}
				err := worker.Os.Send(newMessage)
				if err != nil {
					panic(err)
				}

			}
		}
	}
}

func informReceiverTaskStatus(worker *Worker, t *TaskInfo, content string) {
	newMessage := Message{
		From:    worker.GetHostName(),
		To:      worker.Manager,
		Content: content,
		Body:    *t,
	}
	err := worker.Os.Send(newMessage)
	if err != nil {
		panic(err)
	}
}

func _calculateNodeInfo(worker *Worker) NodeInfo {
	var cpu int32 = 0
	var memory int32 = 0

	for _, t := range worker.TaskMap {
		cpu += t.CpuRequest
		memory += t.MemoryRequest
	}

	var nodeinfo NodeInfo = NodeInfo{
		Addr:           worker.GetHostName(),
		Cpu:            config.Val.NodeCpu,
		Memory:         config.Val.NodeMemory,
		CpuAllocted:    cpu,
		MemoryAllocted: memory,
	}
	return nodeinfo
}
