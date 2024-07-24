package base

import (
	"fmt"
	"os/exec"
	"simds-standalone/config"
	"simds-standalone/core"
	"time"
)

type Worker struct {
	core.BasicNode
	Node    core.NodeInfo // do not store the information , calculate when needed from tasks
	Manager string        // will nofify the ther worker's Manager if the task state is changed
	TaskMap map[string]*core.TaskInfo
}

func NewWorker(host string, nodeinfo core.NodeInfo, manager string) *Worker {
	return &Worker{
		BasicNode: core.BasicNode{Host: host},
		Manager:   manager,
		Node:      nodeinfo,
		TaskMap:   map[string]*core.TaskInfo{},
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

func (worker *Worker) Update(msg core.Message) {

	switch msg.Content {

	case "TaskStart":

		taskid := msg.Body.(core.TaskInfo).Id
		if t, ok := worker.TaskMap[taskid]; ok {
			if t.Status == "needStart" {
				worker.deakRunTask(*t)
			}
		}
	case "TaskPreAllocate":
		newTask := msg.Body.(core.TaskInfo)
		worker.TaskMap[newTask.Id] = &newTask
		newTask.Status = "needStart"
	case "TaskRun":
		newTask := msg.Body.(core.TaskInfo)
		worker.deakRunTask(newTask)

	case "TaskCancelAlloc":
		taskid := msg.Body.(core.TaskInfo).Id
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
		t := msg.Body.(core.TaskInfo)
		id := t.Id
		if worker.Manager != "" {
			informReceiverTaskStatus(worker, &t, "TaskFinish")
		}
		delete(worker.TaskMap, id)
		nodeinfo := _calculateNodeInfo(worker)
		worker.Os.LogInfo("stdout", worker.GetHostName(), "TaskFinish", fmt.Sprint(nodeinfo))
		worker.Os.Send(core.Message{
			From:    worker.Host,
			To:      t.User,
			Content: "TaskFinish",
			Body:    t,
		})
	}

}

func (worker *Worker) deakRunTask(t core.TaskInfo) {
	worker.TaskMap[t.Id] = &t
	t.StartTime = worker.Os.GetTime()
	t.LeftTime = t.LifeTime
	t.Status = "start"
	worker._runTask(t)
	worker.Os.Send(core.Message{
		From:    worker.Host,
		To:      t.User,
		Content: "TaskStart",
		Body:    t,
	})
}

func (node *Worker) _runTask(t core.TaskInfo) {
	node.Os.Run(func() {
		cmd := exec.Command("bash", "-c", t.Cmd)
		err := cmd.Run()
		if err != nil {
			panic(err)
		}
		newMessage := core.Message{
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
				newMessage := core.Message{
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

func informReceiverTaskStatus(worker *Worker, t *core.TaskInfo, content string) {
	newMessage := core.Message{
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

func _calculateNodeInfo(worker *Worker) core.NodeInfo {
	var cpu int32 = 0
	var memory int32 = 0

	for _, t := range worker.TaskMap {
		cpu += t.CpuRequest
		memory += t.MemoryRequest
	}

	var nodeinfo core.NodeInfo = core.NodeInfo{
		Addr:           worker.GetHostName(),
		Cpu:            config.Val.NodeCpu,
		Memory:         config.Val.NodeMemory,
		CpuAllocted:    cpu,
		MemoryAllocted: memory,
	}
	return nodeinfo
}
