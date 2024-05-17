package core

import (
	"fmt"
	"simds-standalone/common"
	"simds-standalone/config"
	"time"
)

const TASKS_EVENT_LOG_NAME = "tasks_event.log"

type TaskGen struct {
	BasicNode
	Started   bool
	StartTime time.Time
	CurTaskId int
	Receivers []string
	Src       []SrcNode
}

type SrcNode struct {
	time time.Duration
	task TaskInfo
}

// NewTaskGen 创造空的TaskGen
func NewTaskGen(hostname string) *TaskGen {

	taskgen := &TaskGen{
		BasicNode: BasicNode{Host: hostname},
		CurTaskId: 0,
	}

	switch config.Val.TaskMode {
	case "onePeak":
		taskgen.Src = onePeakTaskStream()
	case "noWave":
		taskgen.Src = noWaveTaskStream()
	case "trace":
		src := readTraceTaskStream(config.Val.TraceFile, 1.0, config.Val.SimulateDuration-10000)
		src = applyLoadRate(src, float64(config.Val.NodeNum)/float64(1000)*float64(config.Val.TaskNumFactor)/7.0)
		taskgen.Src = src
	default:
		panic("this mod is not implented")
	}

	return taskgen
}

// 负载没有波动的连续任务流
func noWaveTaskStream() []SrcNode {
	taskNumPerSecond := config.Val.TaskNumFactor * float32(config.Val.NodeNum)
	var sendDuration = time.Duration(config.Val.SimulateDuration-10000) * time.Millisecond
	allTasksNum := int(float32(sendDuration/time.Second) * taskNumPerSecond)
	src := make([]SrcNode, 0, allTasksNum)

	for i := 0; i < allTasksNum; i++ {
		newTask := TaskInfo{
			Id:            fmt.Sprintf("task%d", i),
			CpuRequest:    common.RandIntWithRange(config.Val.TaskCpu, 0.5),
			MemoryRequest: common.RandIntWithRange(config.Val.TaskMemory, 0.5),
			LifeTime:      time.Duration(common.RandIntWithRange(config.Val.TaskLifeTime, 0.5)) * time.Millisecond,
			Status:        "submit",
		}

		t := time.Duration(int64(i) * int64(sendDuration) / int64(allTasksNum))

		src = append(src, SrcNode{t, newTask})

	}
	return src
}

// 有一个峰值的连续任务流
func onePeakTaskStream() []SrcNode {
	taskNumPerSecond := config.Val.TaskNumFactor * float32(config.Val.NodeNum)
	baseTimeDelta := int64(time.Second) / int64(taskNumPerSecond)
	src := make([]SrcNode, 0)
	for i := 0; ; i++ {
		lifeTime := time.Duration(common.RandIntWithRange(config.Val.TaskLifeTime, 0.5)) * time.Millisecond
		newTask := TaskInfo{
			Id:            fmt.Sprintf("task%d", i),
			CpuRequest:    common.RandIntWithRange(config.Val.TaskCpu, 0.5),
			MemoryRequest: common.RandIntWithRange(config.Val.TaskMemory, 0.5),
			LifeTime:      lifeTime,
			Status:        "submit",
			Cmd:           fmt.Sprintf("sleep %f", lifeTime.Seconds()),
		}

		var t time.Duration

		var sendDuration = time.Duration(config.Val.SimulateDuration-10000) * time.Millisecond

		if i == 0 {
			t = time.Duration(0)
		} else if src[i-1].time < sendDuration*2/10 {
			t = src[i-1].time + time.Duration(baseTimeDelta*3/2)
		} else if src[i-1].time < sendDuration*8/10 {
			t = src[i-1].time + time.Duration(baseTimeDelta*3/4)
		} else if src[i-1].time < sendDuration {
			t = src[i-1].time + time.Duration(baseTimeDelta*3/2)
		} else {
			break
		}

		src = append(src, SrcNode{time.Duration(t), newTask})
	}
	return src
}

func (n *TaskGen) Debug() {}

func (taskgen *TaskGen) Update() {
	if !taskgen.Started {
		taskgen.StartTime = taskgen.Os.GetTime()
		taskgen.Started = true
		return
	}

	for taskgen.Os.HasMessage() {
		msg, err := taskgen.Os.Recv()
		if err != nil {
			panic(err)
		}

		switch msg.Content {
		case "TaskStart":
			newtask := msg.Body.(TaskInfo)
			taskgen.Os.LogInfo(TASKS_EVENT_LOG_NAME, newtask.Id, "start", msg.From, fmt.Sprint(newtask.CpuRequest), fmt.Sprint(newtask.MemoryRequest))
		case "TaskFinish":
			newtask := msg.Body.(TaskInfo)
			taskgen.Os.LogInfo(TASKS_EVENT_LOG_NAME, newtask.Id, "finish", msg.From, fmt.Sprint(newtask.CpuRequest), fmt.Sprint(newtask.MemoryRequest))
		case "TaskCommitFail":
			task := msg.Body.(TaskInfo)
			newMessage := Message{
				From:    taskgen.GetHostName(),
				To:      msg.From,
				Content: "TaskDispense",
				Body:    task,
			}
			err := taskgen.Os.Send(newMessage)
			if err != nil {
				panic(err)
			}
		}
	}

	taskgenAddr := taskgen.GetHostName()

	receiverNum := len(taskgen.Receivers)

	timeNow := taskgen.Os.GetTime().Sub(taskgen.StartTime)
	for taskgen.CurTaskId < len(taskgen.Src) {
		if taskgen.Src[taskgen.CurTaskId].time > timeNow {
			break
		}

		newtask := taskgen.Src[taskgen.CurTaskId].task
		newtask.User = taskgen.Host
		receiverAddr := taskgen.Receivers[taskgen.CurTaskId%receiverNum]
		newMessage := Message{
			From:    taskgenAddr,
			To:      receiverAddr,
			Content: "TaskDispense",
			Body:    newtask,
		}
		err := taskgen.Os.Send(newMessage)
		if err != nil {
			panic(err)
		}
		taskgen.Os.LogInfo(TASKS_EVENT_LOG_NAME, newtask.Id, "submit", receiverAddr, fmt.Sprint(newtask.CpuRequest), fmt.Sprint(newtask.MemoryRequest))
		taskgen.CurTaskId++
	}
}

func (taskgen *TaskGen) SimulateTasksUpdate() {
}
