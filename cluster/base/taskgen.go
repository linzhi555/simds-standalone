package base

import (
	"fmt"
	"log"
	"simds-standalone/common"
	"simds-standalone/config"
	"strings"
	"time"
)

type TaskGen struct {
	BasicActor
	Status    string
	StartTime time.Time
	CurTaskId int
	Receivers []string
	Src       []SrcNode
}

type SrcNode struct {
	time time.Duration
	task TaskInfo
}

func ConcateStream(a, b []SrcNode, gap time.Duration) []SrcNode {
	var res []SrcNode = make([]SrcNode, 0, len(a)+len(b))
	res = append(res, a...)
	res = append(res, b...)

	lastATime := a[len(a)-1].time
	for i := len(a); i < len(a)+len(b); i++ {
		res[i].time += (gap + lastATime)
	}
	return res
}

func CutStream(old []SrcNode, until time.Duration) []SrcNode {
	var res []SrcNode
	for i := range old {
		if old[i].time > until {
			break
		}
		res = append(res, old[i])
	}
	return res
}

// NewTaskGen 创造空的TaskGen
func NewTaskGen(hostname string) *TaskGen {

	taskgen := &TaskGen{
		BasicActor: BasicActor{Host: hostname},
		CurTaskId:  0,
	}

	return taskgen
}

func (taskgen *TaskGen) InitTaskSrc(src []SrcNode) {
	taskgen.Src = src
}

func testTaskStream() []SrcNode {
	switch config.Val.TaskMode {
	case "onePeak":
		return onePeakTaskStream()
	case "noWave":
		return noWaveTaskStream()
	case "trace":
		src := readTraceTaskStream(config.Val.TraceFile, 1.0, config.Val.SimulateDuration-10000)
		src = applyLoadRate(src, float64(config.Val.NodeNum)/float64(1000)*float64(config.Val.TaskNumFactor)/7.0)
		return src
	}
	panic("error task stream type")
}

func preheatTaskStream() []SrcNode {
	src := noWaveTaskStream()
	for i := range src {
		src[i].task.Id += "_preheat"
	}
	return CutStream(src, 4*time.Second)
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

func (taskgen *TaskGen) Update(msg Message) {
	switch msg.Content {
	case "SignalBoot":
		formal := testTaskStream()
		preheat := preheatTaskStream() // before formal test, there is a preheat stream to warm up the system.
		all := ConcateStream(preheat, formal, 5*time.Second)

		taskgen.InitTaskSrc(all)
		taskgen.StartTime = taskgen.Os.GetTime()
		taskgen.Status = "preheat"
		taskgen.Os.Run(func() { taskgen._sendingTask() })
	case "TaskStart":
		newtask := msg.Body.(TaskInfo)
		if !strings.HasSuffix(newtask.Id, "preheat") {
			taskgen.Status = "formalTest"
		}

	case "TaskFinish":

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

func (taskgen *TaskGen) _sendingTask() {
	taskgenAddr := taskgen.GetHostName()

	receiverNum := len(taskgen.Receivers)
	log.Println("start sending task")

	for taskgen.CurTaskId < len(taskgen.Src) {
		for taskgen.Src[taskgen.CurTaskId].time > taskgen.Os.GetTime().Sub(taskgen.StartTime) {
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

		taskgen.CurTaskId++
	}
}

func (taskgen *TaskGen) SimulateTasksUpdate() {
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

		taskgen.CurTaskId++
	}
}
