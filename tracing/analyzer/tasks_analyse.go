package analyzer

import (
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"time"

	"simds-standalone/cluster/base"
	"simds-standalone/common"
	"simds-standalone/config"
)

func AnalyseTasks(taskLogFile string, outdir string) {
	events := ReadTaskEventCsv(taskLogFile)
	events.Output(outdir, "all_events.log")
	events.OutputTaskSubmitRate(outdir)

	AnalyzeEventRate(events, SUBMIT, 100).Output(outdir, "taskSubmit")
	AnalyzeStageDuration(events, START, FINISH).Output(outdir, "lifeTime")
	latencies := AnalyzeStageDuration(events, SUBMIT, START)
	latencies.Output(outdir, "latency")

	InitCluster(events, latencies).ReplayEvents().Output(outdir)
}

var (
	SUBMIT = "TaskDispense"
	START  = "TaskStart"
	FINISH = "TaskFinish"
)

var FAIL = 999 * time.Hour

var TASK_EVENT_LOG_HEAD = []string{"time", "type", "taskid", "actorid", "cpu", "memory"}

type TaskEvent struct {
	Time    time.Time
	Type    string
	TaskId  string
	ActorId string
	Cpu     int32
	Memory  int32
}

const (
	_TTime = iota
	_TType
	_TTaskId
	_TActorId
	_TCpu
	_TMemory
)

type TaskEventLine []*TaskEvent

// for sort
func (l TaskEventLine) Len() int      { return len(l) }
func (l TaskEventLine) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l TaskEventLine) Less(i, j int) bool {
	return l[i].Time.Before(l[j].Time)
}

// for EventLines interface
func (l TaskEventLine) GetID(i int) string            { return l[i].TaskId }
func (l TaskEventLine) GetType(i int) string          { return l[i].Type }
func (l TaskEventLine) GetHappenTime(i int) time.Time { return l[i].Time }

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
func (l TaskEventLine) Output(outputDir string, filename string) {
	outputlogfile := path.Join(outputDir, filename)

	err := common.AppendLineCsvFile(outputlogfile, TASK_EVENT_LOG_HEAD)
	if err != nil {
		panic(err)
	}

	for _, event := range l {
		err := common.AppendLineCsvFile(outputlogfile, event.Strings())
		if err != nil {
			panic(err)
		}
	}
}

func (l TaskEventLine) OutputTaskSubmitRate(outputDir string) {
	outputfile := path.Join(outputDir, "task_speed.log")
	startTime := l[0].Time
	for _, event := range l {

		if event.Type == SUBMIT {
			err := common.AppendLineCsvFile(outputfile, []string{fmt.Sprint(event.Time.Sub(startTime).Milliseconds()), event.Type})
			if err != nil {
				panic(err)
			}
		}

	}
}

func strings2TaskEvent(line []string) *TaskEvent {
	var t TaskEvent
	time, err := time.Parse(time.RFC3339Nano, line[_TTime])
	if err != nil {
		panic(err)
	}

	t.Time = time
	t.TaskId = line[_TTaskId]
	t.Type = line[_TType]
	t.ActorId = line[_TActorId]

	cpu := common.Str_to_int64(line[_TCpu])

	t.Cpu = int32(cpu)

	memory := common.Str_to_int64(line[_TMemory])
	t.Memory = int32(memory)
	return &t
}

func (t *TaskEvent) Strings() (line []string) {
	line = make([]string, 6)
	line[_TTime] = t.Time.Format(time.RFC3339Nano)
	line[_TTaskId] = t.TaskId
	line[_TType] = string(t.Type)
	line[_TActorId] = t.ActorId
	line[_TCpu] = fmt.Sprint(t.Cpu)
	line[_TMemory] = fmt.Sprint(t.Memory)
	return
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
	if common.IsFileExist(outputlogfile) {
		err := os.Remove(outputlogfile)
		if err != nil {
			log.Println(err)
		}

	}
	err := common.AppendLineCsvFile(outputlogfile, []string{"time_ms", "taskLatency", "cpuAvg", "ramAvg", "memVar", "memVar"})
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
	TaskLatencyList StageCostList
	Nodes           map[string]*base.NodeInfo
}

// create a new cluster according the taskevent logs.
// if the ip appears in the logs then create this node.
func InitCluster(events TaskEventLine, latencies StageCostList) *Cluster {
	var cluster Cluster
	cluster.AllEvents = events
	cluster.TaskLatencyList = latencies
	cluster.Nodes = make(map[string]*base.NodeInfo)

	for _, event := range events {
		if event.Type != START {
			continue
		}

		if _, ok := cluster.Nodes[event.ActorId]; ok == true {
			continue
		}

		cluster.Nodes[event.ActorId] = &base.NodeInfo{
			Addr:           event.ActorId,
			Cpu:            config.Val.NodeCpu,
			Memory:         config.Val.NodeMemory,
			CpuAllocted:    0,
			MemoryAllocted: 0,
		}
	}
	return &cluster
}

func (c *Cluster) ReplayEvents() (statusLine ClusterStatusLine) {
	lastRecordTime := c.AllEvents[0].Time
	taskAfterLastRecord := []string{}

	taskLatencyMap := map[string]time.Duration{}
	for _, record := range c.TaskLatencyList {
		taskLatencyMap[record.Id] = record.Cost
	}

	for _, e := range c.AllEvents {
		switch e.Type {
		case SUBMIT:
			taskAfterLastRecord = append(taskAfterLastRecord, e.TaskId)
		case START:
			c.Nodes[e.ActorId].AddAllocated(e.Cpu, e.Memory)
		case FINISH:
			c.Nodes[e.ActorId].SubAllocated(e.Cpu, e.Memory)
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

			statusLine = append(statusLine, ClusterStatus{e.Time, latencyMax, c.calMetrics()})
			taskAfterLastRecord = []string{}
			lastRecordTime = e.Time
		}
	}

	return statusLine

}

func (c *Cluster) calMetrics() ClusterMetric {
	var allCpu, allRam int32 = 0, 0
	for _, node := range c.Nodes {
		allCpu += node.CpuAllocted
		allRam += node.MemoryAllocted
	}

	averageCpuUsedPer := float32(allCpu) / float32(int32(len(c.Nodes))*config.Val.NodeCpu)
	averageRamUsedPer := float32(allRam) / float32(int32(len(c.Nodes))*config.Val.NodeMemory)

	var cpuUsedPerVariance, ramUsedPerVariance float32 = 0.0, 0.0
	for _, node := range c.Nodes {
		cpuUsedPerVariance += pow2(node.CpuPercent() - averageCpuUsedPer)
		ramUsedPerVariance += pow2(node.MemoryPercent() - averageRamUsedPer)
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