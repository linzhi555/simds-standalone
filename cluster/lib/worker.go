package lib

import (
	//"os/exec"
	//"time"

	"simds-standalone/cluster/base"
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
				worker.dealRunTask(*t)
			}
		}
	case "TaskPreAllocate":
		newTask := msg.Body.(TaskInfo)
		worker.TaskMap[newTask.Id] = &newTask
		newTask.Status = "needStart"
	case "TaskRun":
		newTask := msg.Body.(TaskInfo)
		worker.dealRunTask(newTask)

	case "TaskCancelAlloc":
		taskid := msg.Body.(TaskInfo).Id
		if t, ok := worker.TaskMap[taskid]; ok {
			if t.Status == "needStart" {
				t.Status = "finish"
				if worker.Manager != "" {
					informReceiverTaskStatus(worker, t, "TaskCancelled")
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

func (worker *Worker) dealRunTask(t TaskInfo) {
	worker.TaskMap[t.Id] = &t
	t.StartTime = worker.Os.GetTime()
	t.LeftTime = t.LifeTime

	worker.Os.RunCmd(func(err error) {
		worker.Os.Send(base.Message{
			From: worker.GetAddress(),
			To:   worker.GetAddress(),
			Head: "TaskFinish",
			Body: *(t.Clone()),
		})
	}, t.Cmd)

	worker.Os.Send(base.Message{
		From: worker.Host,
		To:   t.User,
		Head: "TaskStart",
		Body: *(t.Clone()),
	})
}

func informReceiverTaskStatus(worker *Worker, t *TaskInfo, content string) {
	worker.Os.Send(base.Message{
		From: worker.GetAddress(),
		To:   worker.Manager,
		Head: content,
		Body: *(t.Clone()),
	})

}
