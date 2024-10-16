package analyzer

import (
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"simds-standalone/cluster/lib"
	"simds-standalone/common"
	"simds-standalone/config"
)

func AnalyseTasks(taskLogFile string, outdir string) {
	events := ReadTaskEventCsv(taskLogFile)
	//events.Output(outdir, "sorted_events.log")

	AnalyzeEventRate(events, SUBMIT, 100).Output(outdir, "_taskSubmit")
	AnalyzeStageDuration(events, START, FINISH).Output(outdir, "_lifeTime")
	latencies := AnalyzeStageDuration(events, SUBMIT, START)
	latencies.Output(outdir, "_taskLatency")

	InitCluster(events).ReplayEvents().Output(outdir, "_clusterStatus")
}

const (
	SUBMIT = "TaskDispense"
	START  = "TaskStart"
	FINISH = "TaskFinish"
)

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
	table, _, err := common.CsvToList(csvfilePath)
	if err != nil {
		if strings.HasPrefix(err.Error(), "partial error:") {
			log.Println(err)
		} else {
			panic(err)
		}
	}
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

func strings2TaskEvent(line []string) *TaskEvent {
	var t TaskEvent
	time, err := common.ParseTime(line[_TTime])
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
	Time   time.Time
	Metric ClusterMetric
}

func (status *ClusterStatus) Strings(startTime time.Time) []string {
	return append([]string{fmt.Sprint(status.Time.Sub(startTime).Milliseconds())}, status.Metric.Strings()...)
}

type ClusterStatusLine []ClusterStatus

func (l ClusterStatusLine) Output(outputDir string, filename string) {
	outputlogfile := path.Join(outputDir, filename+".log")
	if common.IsFileExist(outputlogfile) {
		err := os.Remove(outputlogfile)
		if err != nil {
			log.Println(err)
		}

	}
	err := common.AppendLineCsvFile(outputlogfile, []string{"time_ms", "cpuAvg", "ramAvg", "memVar", "memVar"})
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
	AllEvents TaskEventLine
	Nodes     map[string]*lib.NodeInfo
}

// create a new cluster according the taskevent logs.
// if the ip appears in the logs then create this node.
func InitCluster(events TaskEventLine) *Cluster {
	var cluster Cluster
	cluster.AllEvents = events
	cluster.Nodes = make(map[string]*lib.NodeInfo)

	for _, event := range events {
		if event.Type != START {
			continue
		}

		if _, ok := cluster.Nodes[event.ActorId]; ok {
			continue
		}

		cluster.Nodes[event.ActorId] = &lib.NodeInfo{
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

	for _, e := range c.AllEvents {
		switch e.Type {

		case START:
			c.Nodes[e.ActorId].AddAllocated(e.Cpu, e.Memory)
		case FINISH:
			c.Nodes[e.ActorId].SubAllocated(e.Cpu, e.Memory)
		}
		if e.Time.Sub(lastRecordTime) >= time.Millisecond*100 {
			statusLine = append(statusLine, ClusterStatus{e.Time, c.calMetrics()})
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
