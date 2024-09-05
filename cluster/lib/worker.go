package lib

import (
	"os/exec"
	"time"

	"simds-standalone/cluster/base"
	"simds-standalone/config"
)

type Worker struct {
	base.BasicActor
	Node    NodeInfo // do not store the information , calculate when needed from tasks
	Manager string   // will nofify the ther worker's Manager if the task state is changed
	TaskMap map[string]*TaskInfo
}

func NewWorker(host string, nodeinfo NodeInfo, manager string) *Worker {
	return &Worker{
		BasicActor: base.BasicActor{Host: host},
		Manager:    manager,
		Node:       nodeinfo,
		TaskMap:    map[string]*TaskInfo{},
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

func (worker *Worker) Update(msg base.Message) {

	switch msg.Head {

	case "TaskStart":

		taskid := msg.Body.(TaskInfo).Id
		if t, ok := worker.TaskMap[taskid]; ok {
			if t.Status == "needStart" {
				worker.deakRunTask(*t)
			}
		}
	case "TaskPreAllocate":
		newTask := msg.Body.(TaskInfo)
		worker.TaskMap[newTask.Id] = &newTask
		newTask.Status = "needStart"
	case "TaskRun":
		newTask := msg.Body.(TaskInfo)
		worker.deakRunTask(newTask)

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
		worker.Os.Send(base.Message{
			From: worker.Host,
			To:   t.User,
			Head: "TaskFinish",
			Body: t,
		})
	}

}

func (worker *Worker) deakRunTask(t TaskInfo) {
	worker.TaskMap[t.Id] = &t
	t.StartTime = worker.Os.GetTime()
	t.LeftTime = t.LifeTime
	t.Status = "start"
	worker._runTask(t)
	worker.Os.Send(base.Message{
		From: worker.Host,
		To:   t.User,
		Head: "TaskStart",
		Body: t,
	})
}

func (node *Worker) _runTask(t TaskInfo) {
	node.Os.Run(func() {
		cmd := exec.Command("bash", "-c", t.Cmd)
		err := cmd.Run()
		if err != nil {
			panic(err)
		}
		newMessage := base.Message{
			From: node.GetAddress(),
			To:   node.GetAddress(),
			Head: "TaskFinish",
			Body: t,
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
				newMessage := base.Message{
					From: worker.GetAddress(),
					To:   worker.GetAddress(),
					Head: "TaskFinish",
					Body: *t,
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
	newMessage := base.Message{
		From: worker.GetAddress(),
		To:   worker.Manager,
		Head: content,
		Body: *t,
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
		Addr:           worker.GetAddress(),
		Cpu:            config.Val.NodeCpu,
		Memory:         config.Val.NodeMemory,
		CpuAllocted:    cpu,
		MemoryAllocted: memory,
	}
	return nodeinfo
}
