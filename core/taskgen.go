package core

import (
	"fmt"
	"simds-standalone/common"
	"simds-standalone/config"
	"time"
)

type TaskGen struct {
	Os        OsApi
	Host      string
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
		Host:      hostname,
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
		newTask := TaskInfo{
			Id:            fmt.Sprintf("task%d", i),
			CpuRequest:    common.RandIntWithRange(config.Val.TaskCpu, 0.5),
			MemoryRequest: common.RandIntWithRange(config.Val.TaskMemory, 0.5),
			LifeTime:      time.Duration(common.RandIntWithRange(config.Val.TaskLifeTime, 0.5)) * time.Millisecond,
			Status:        "submit",
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

// SetOsApi for NodeComponent interface
func (n *TaskGen) SetOsApi(osapi OsApi) { n.Os = osapi }

func (n *TaskGen) Debug() {}

// CommonTaskgenUpdate 通用的任务发生器，适
// 用多种形式集群，通过Receivers字段来确定发送对象
func (taskgen *TaskGen) Update() {
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

		taskgen.Os.LogInfo(taskgen.Os, fmt.Sprintf(": send task to %s %v", receiverAddr, newMessage.Body))
		//TaskEventLog(taskgen.Os.GetTime(), &newtask, receiverAddr)
		//TaskSpeedLog(_getTime_ms(taskgen.Os), &newtask)
		taskgen.CurTaskId++
	}
}
