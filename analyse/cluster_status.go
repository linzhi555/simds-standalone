package main

import (
	"fmt"
	"path"
	"simds-standalone/common"
	"sort"
	"strconv"
	"time"
)

var (
	SUBMIT = "submit"
	START  = "start"
	FINISH = "finish"
)

type TaskEvent struct {
	Time   time.Time
	TaskId string
	Type   string
	NodeIp string
	Cpu    float32
	Ram    float32
}

type TaskEventLine []*TaskEvent

func (l TaskEventLine) Len() int      { return len(l) }
func (l TaskEventLine) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l TaskEventLine) Less(i, j int) bool {
	return l[i].Time.Before(l[j].Time)
}

// read the TaskEvent csv file
func ReadTaskEventCsv(csvfilePath string) TaskEventLine {

	table, _ := common.CsvToList(csvfilePath)

	var eventLine []*TaskEvent
	for _, line := range table {
		eventLine = append(eventLine, strings2TaskEvent(line))
	}
	sort.Sort(TaskEventLine(eventLine))
	return eventLine
}

const (
	_time   = 0
	_taskid = 1
	_type   = 2
	_nodeip = 3
	_cpu    = 4
	_ram    = 5
)

func strings2TaskEvent(line []string) *TaskEvent {
	var t TaskEvent
	time, err := time.Parse(time.RFC3339Nano, line[_time])
	if err != nil {
		panic(err)
	}

	t.Time = time
	t.TaskId = line[_taskid]
	t.Type = line[_type]
	t.NodeIp = line[_nodeip]

	cpu, err := strconv.ParseFloat(line[_cpu], 32)
	if err != nil {
		panic(err)
	}
	t.Cpu = float32(cpu)

	ram, err := strconv.ParseFloat(line[_ram], 32)
	if err != nil {
		panic(err)
	}
	t.Ram = float32(ram)

	return &t
}
func (t *TaskEvent) Strings() (line []string) {
	line = make([]string, 6)
	line[_time] = t.Time.Format(time.RFC3339Nano)
	line[_taskid] = t.TaskId
	line[_type] = string(t.Type)
	line[_nodeip] = t.NodeIp
	line[_cpu] = fmt.Sprint(t.Cpu)
	line[_ram] = fmt.Sprint(t.Ram)
	return
}

type TaskStageCostList []struct {
	Taskid string
	Cost   time.Duration
}

func (l TaskStageCostList) Len() int      { return len(l) }
func (l TaskStageCostList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l TaskStageCostList) Less(i, j int) bool {
	return l[i].Cost.Nanoseconds() < l[j].Cost.Nanoseconds()
}

func (l TaskEventLine) AnalyseStageDuration(stage1 string, stage2 string) TaskStageCostList {
	eventStage1 := make(map[string]time.Time)
	eventStage2 := make(map[string]time.Time)

	for _, event := range l {
		switch event.Type {
		case stage1:
			eventStage1[event.TaskId] = event.Time
		case stage2:
			eventStage2[event.TaskId] = event.Time
		default:
		}
	}

	var res TaskStageCostList
	for taskid := range eventStage1 {
		var temp struct {
			Taskid string
			Cost   time.Duration
		}
		temp.Taskid = taskid
		if _, ok := eventStage2[taskid]; ok {
			temp.Cost = eventStage2[taskid].Sub(eventStage1[taskid])
		} else {
			temp.Cost = 999 * time.Hour
		}
		res = append(res, temp)

	}
	sort.Sort(res)
	return res
}

func (l TaskEventLine) AnalyseSchedulerLatency(outPutDir string) {
	costList := l.AnalyseStageDuration(SUBMIT, START)
	outPutFigurePath := path.Join(outPutDir, "latencyCurve.png")
	outPutLogPath := path.Join(outPutDir, "latencyCurve.log")
	outPutMetricPath := path.Join(outPutDir, "latency_metric.log")

	err := common.AppendLineCsvFile(outPutLogPath, []string{"taskid", "latency"})
	if err != nil {
		panic(err)
	}

	for _, line := range costList {
		err := common.AppendLineCsvFile(outPutLogPath, []string{fmt.Sprint(line.Taskid), fmt.Sprint(line.Cost)})
		if err != nil {
			panic(err)
		}
	}

	var sum time.Duration = 0
	for _, line := range costList {
		sum += line.Cost
	}
	average := time.Duration(int64(sum) / int64(len(costList)))
	high_90_per := costList[(len(costList)*9)/10].Cost
	high_99_per := costList[(len(costList)*99)/100].Cost
	err = common.AppendLineCsvFile(outPutMetricPath, []string{"average", "90%high", "99%high"})
	if err != nil {
		panic(err)
	}

	err = common.AppendLineCsvFile(outPutMetricPath, []string{fmt.Sprint(average), fmt.Sprint(high_90_per), fmt.Sprint(high_99_per)})
	if err != nil {
		panic(err)
	}

	outputLatencyResultFigure(outPutFigurePath, costList)
}

func (l TaskEventLine) AnalyseTaskLifeTime(outPutDir string) {
	costList := l.AnalyseStageDuration(START, FINISH)
	outPutLogPath := path.Join(outPutDir, "lifeTimeCurve.log")
	outPutMetricPath := path.Join(outPutDir, "lifeTime_metric.log")

	err := common.AppendLineCsvFile(outPutLogPath, []string{"taskid", "lifetime"})
	if err != nil {
		panic(err)
	}

	for _, line := range costList {
		err := common.AppendLineCsvFile(outPutLogPath, []string{fmt.Sprint(line.Taskid), fmt.Sprint(line.Cost)})
		if err != nil {
			panic(err)
		}
	}

	var sum time.Duration = 0
	for _, line := range costList {
		sum += line.Cost
	}
	average := time.Duration(int64(sum) / int64(len(costList)))
	high_90_per := costList[(len(costList)*9)/10].Cost
	high_99_per := costList[(len(costList)*99)/100].Cost

	err = common.AppendLineCsvFile(outPutMetricPath, []string{"average", "90%high", "99%high"})
	if err != nil {
		panic(err)
	}
	err = common.AppendLineCsvFile(outPutMetricPath, []string{fmt.Sprint(average), fmt.Sprint(high_90_per), fmt.Sprint(high_99_per)})

	if err != nil {
		panic(err)
	}
}

func AdjustEventTimeByTimeRate(timeRate int, events TaskEventLine) {
	zeroTime := events[0].Time
	for i := range events {
		events[i].Time = zeroTime.Add(events[i].Time.Sub(zeroTime) / time.Duration(timeRate))
	}
}

type Node struct {
	Ip         string
	CpuCap     float32 //cpu capacity
	RamCap     float32 //ram capacity
	CpuUsed    float32
	RamUsed    float32
	CpuUsedPer float32 // Cpu Usage percentage
	RamUsedPer float32
}

func (n *Node) AddTask(cpu float32, ram float32) {
	n.CpuUsed += cpu
	n.RamUsed += ram

	n.CpuUsedPer = n.CpuUsed / n.CpuCap
	n.RamUsedPer = n.RamUsed / n.RamCap
}

func (n *Node) RemoveTask(cpu float32, ram float32) {
	n.CpuUsed -= cpu
	n.RamUsed -= ram

	//	if n.CpuUsed <0.0{
	//		n.CpuUsed = 0.0
	//	}
	//	if n.RamUsed <0.0{
	//		n.RamUsed = 0.0
	//	}

	n.CpuUsedPer = n.CpuUsed / n.CpuCap
	n.RamUsedPer = n.RamUsed / n.RamCap
}

type ClusterMetric struct {
	CpuUsedPerAverage  float32
	RamUsedPerAverage  float32
	CpuUsedPerVariance float32
	RamUsedPerVariance float32
}

type ClusterStatus struct {
	Time   time.Time
	Metric ClusterMetric
}

type Cluster map[string]*Node

// create a new cluster according the taskevent logs.
// if the ip appears in the logs then create this node.
func InitCluster(l TaskEventLine) Cluster {
	var cluster Cluster = make(Cluster)

	for _, event := range l {
		if event.Type != START {
			continue
		}
		if _, ok := cluster[event.NodeIp]; ok == true {
			continue
		}
		cluster[event.NodeIp] = &Node{event.NodeIp, 10.0, 10.0, 0.0, 0.0, 0.0, 0.0}
	}

	return cluster
}

func (c Cluster) CalMetrics() ClusterMetric {
	var allCpu, allRam float32 = 0.0, 0.0
	var cpuUsedPerVariance, ramUsedPerVariance float32 = 0.0, 0.0
	for _, node := range c {
		allCpu += node.CpuUsedPer
		allRam += node.RamUsedPer
	}
	averageCpuUsedPer := allCpu / float32(len(c))
	averageRamUsedPer := allRam / float32(len(c))

	for _, node := range c {
		cpuUsedPerVariance += pow2(node.CpuUsedPer - averageCpuUsedPer)
		ramUsedPerVariance += pow2(node.RamUsedPer - averageRamUsedPer)
	}
	cpuUsedPerVariance /= float32(len(c))
	ramUsedPerVariance /= float32(len(c))

	return ClusterMetric{
		averageCpuUsedPer,
		averageRamUsedPer,
		cpuUsedPerVariance,
		ramUsedPerVariance,
	}
}

func pow2(a float32) float32 {
	return a * a
}

func (c Cluster) CalStatusCurves(events []*TaskEvent) (statusLine []ClusterStatus) {
	for _, e := range events {
		switch e.Type {
		case START:
			c[e.NodeIp].AddTask(e.Cpu, e.Ram)
		case FINISH:
			c[e.NodeIp].RemoveTask(e.Cpu, e.Ram)
		}
		statusLine = append(statusLine, ClusterStatus{e.Time, c.CalMetrics()})
	}
	return
}
