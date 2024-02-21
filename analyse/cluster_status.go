package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"time"

	"simds-standalone/common"
	"simds-standalone/config"
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

// outout eventsLine to csv
func (l TaskEventLine) Output(outputDir string) {
	outputlogfile := path.Join(outputDir, "all_events.log")
	err := common.AppendLineCsvFile(outputlogfile, []string{"time", "taskid", "type", "nodeip", "cpu", "ram"})
	if err != nil {
		panic(err)
	}

	for _, event := range l {
		err = common.AppendLineCsvFile(outputlogfile, event.Strings())
		if err != nil {
			panic(err)
		}
	}
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

type Node struct {
	Ip         string
	CpuCap     float32 //cpu capacity
	RamCap     float32 //ram capacity
	Tasks      map[string]struct{}
	CpuUsedPer float32
	RamUsedPer float32
}

func NewNode(ip string, cpu, ram float32) *Node {
	return &Node{
		Ip:         ip,
		CpuCap:     cpu,
		RamCap:     ram,
		Tasks:      map[string]struct{}{},
		CpuUsedPer: 0,
		RamUsedPer: 0,
	}
}

func (n *Node) AddTask(taskid string) {
	n.Tasks[taskid] = struct{}{}
}

func (n *Node) RemoveTask(taskid string) {
	delete(n.Tasks, taskid)
}

type ClusterMetric struct {
	CpuUsedPerAverage  float32
	RamUsedPerAverage  float32
	CpuUsedPerVariance float32
	RamUsedPerVariance float32
}

func (m *ClusterMetric) Strings() []string {
	return []string{
		fmt.Sprint(m.CpuUsedPerAverage),
		fmt.Sprint(m.RamUsedPerAverage),
		fmt.Sprint(m.CpuUsedPerVariance),
		fmt.Sprint(m.RamUsedPerVariance),
	}
}

type ClusterStatus struct {
	Time        time.Time
	TaskLatency time.Duration
	Metric      ClusterMetric
}

func (status *ClusterStatus) Strings(startTime time.Time) []string {
	return append([]string{fmt.Sprint(status.Time.Sub(startTime).Milliseconds()), fmt.Sprint(status.TaskLatency)}, status.Metric.Strings()...)
}

type ClusterStatusLine []ClusterStatus

func (l ClusterStatusLine) Output(outputDir string) {
	outputlogfile := path.Join(outputDir, "cluster_status.log")
	err := os.Remove(outputlogfile)
	if err != nil {
		log.Println(err)
	}
	err = common.AppendLineCsvFile(outputlogfile, []string{"time_ms", "taskLatency", "cpuAvg", "ramAvg", "cpuVar", "ramVar"})
	if err != nil {
		panic(err)
	}
	startTime := l[0].Time
	for _, status := range l {
		err = common.AppendLineCsvFile(outputlogfile, status.Strings(startTime))
		if err != nil {
			panic(err)
		}
	}
}

type Cluster struct {
	AllEvents       TaskEventLine
	SubmitEvents    map[string]*TaskEvent
	TaskLatencyList TaskStageCostList
	Nodes           map[string]*Node
}

// create a new cluster according the taskevent logs.
// if the ip appears in the logs then create this node.
func InitCluster(l TaskEventLine) *Cluster {
	var cluster Cluster
	cluster.AllEvents = l
	cluster.Nodes = make(map[string]*Node)
	cluster.SubmitEvents = make(map[string]*TaskEvent)

	for _, event := range l {
		if event.Type == SUBMIT {
			cluster.SubmitEvents[event.TaskId] = event
			continue
		}

		if event.Type != START {
			continue
		}
		if _, ok := cluster.Nodes[event.NodeIp]; ok == true {
			continue
		}

		cluster.Nodes[event.NodeIp] = NewNode(event.NodeIp, float32(config.Val.NodeCpu), float32(config.Val.NodeMemory))
	}

	return &cluster
}

func (c *Cluster) CalMetrics() ClusterMetric {
	for _, node := range c.Nodes {
		var cpuUsed, ramUsed float32 = 0, 0
		for t := range node.Tasks {
			cpuUsed += c.SubmitEvents[t].Cpu
			ramUsed += c.SubmitEvents[t].Ram
		}

		node.CpuUsedPer = cpuUsed / node.CpuCap
		node.RamUsedPer = ramUsed / node.RamCap
	}

	var allCpu, allRam float32 = 0.0, 0.0
	for _, node := range c.Nodes {
		allCpu += node.CpuUsedPer
		allRam += node.RamUsedPer
	}
	averageCpuUsedPer := allCpu / float32(len(c.Nodes))
	averageRamUsedPer := allRam / float32(len(c.Nodes))

	var cpuUsedPerVariance, ramUsedPerVariance float32 = 0.0, 0.0
	for _, node := range c.Nodes {
		cpuUsedPerVariance += pow2(node.CpuUsedPer - averageCpuUsedPer)
		ramUsedPerVariance += pow2(node.RamUsedPer - averageRamUsedPer)
	}
	cpuUsedPerVariance /= float32(len(c.Nodes))
	ramUsedPerVariance /= float32(len(c.Nodes))

	return ClusterMetric{
		CpuUsedPerAverage:  averageCpuUsedPer,
		RamUsedPerAverage:  averageRamUsedPer,
		CpuUsedPerVariance: cpuUsedPerVariance,
		RamUsedPerVariance: ramUsedPerVariance,
	}
}

func pow2(a float32) float32 {
	return a * a
}

func (c *Cluster) AnalyseSchedulerLatency(outPutDir string) {
	costList := c.AllEvents.AnalyseStageDuration(SUBMIT, START)
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

	c.TaskLatencyList = costList

	outputLatencyResultFigure(outPutFigurePath, costList)
}

func (c *Cluster) AnalyseTaskLifeTime(outPutDir string) {
	costList := c.AllEvents.AnalyseStageDuration(START, FINISH)
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

func (c *Cluster) CalStatusCurves(outputDir string) (statusLine []ClusterStatus) {
	lastRecordTime := c.AllEvents[0].Time
	taskAfterLastRecord := []string{}

	taskLatencyMap := map[string]time.Duration{}
	for _, record := range c.TaskLatencyList {
		taskLatencyMap[record.Taskid] = record.Cost
	}

	for _, e := range c.AllEvents {
		switch e.Type {
		case SUBMIT:
			taskAfterLastRecord = append(taskAfterLastRecord, e.TaskId)
		case START:
			c.Nodes[e.NodeIp].AddTask(e.TaskId)
		case FINISH:
			c.Nodes[e.NodeIp].RemoveTask(e.TaskId)
		}
		if e.Time.Sub(lastRecordTime) >= time.Millisecond*100 {
			latencyMax := time.Duration(0)
			tasksNum := int64(len(taskAfterLastRecord))

			if tasksNum != 0 {
				for _, task := range taskAfterLastRecord {
					latency := taskLatencyMap[task]
					if latency > latencyMax {
						latencyMax = latency
					}
				}
			}

			statusLine = append(statusLine, ClusterStatus{e.Time, latencyMax, c.CalMetrics()})
			taskAfterLastRecord = []string{}
			lastRecordTime = e.Time
		}
	}

	ClusterStatusLine(statusLine).Output(outputDir)
	return
}
