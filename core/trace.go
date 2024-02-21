package core

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"simds-standalone/common"
)

//const (
//	tasknumCompressRate int     = 4000
//	lifeRate            float64 = 0.0005
//	resourceRate        float64 = 100
//	timebias            int     = 8 // use task after $timebias Second
//)

// read traceFile ,resourceRate: multipy a factor to the resource value
// endTime: how long the trace used, unit milisecond, -1 meaning use all the trace
func readTraceTaskStream(traceFile string, resourceRate float64, endTime int32) []SrcNode {

	src := make([]SrcNode, 0)

	table, _ := common.CsvToList(traceFile)

	for i := range table {
		newTask := TaskInfo{
			Id:            table[i][1], // table [i][1] is the taskid
			CpuRequest:    int32(common.Str_to_float64(table[i][3]) * resourceRate),
			MemoryRequest: int32(common.Str_to_float64(table[i][4]) * resourceRate),
			LifeTime:      time.Duration(common.Str_to_int64(table[i][2])) * time.Microsecond,
			Status:        "submit",
		}
		submitTime := time.Duration(common.Str_to_int64(table[i][0])) * time.Microsecond // table [i][1] is the taskid

		if endTime >= 0 {
			if submitTime > time.Duration(endTime)*time.Millisecond {
				break
			}
		}

		src = append(src, SrcNode{submitTime, newTask})
	}

	log.Println(len(src))
	log.Println(src[0:10])
	return src
}

const tasknumCompressRate = 4000

func DealRawFile(loadrate, lifeRate, resourceRate float64, timebias, maxResourceLimit int32, infile, outfile string) {

	taskTable := readTraceTaskStream(infile, resourceRate, -1)
	startTime := taskTable[0].time

	for i := range taskTable {
		//adjust zero Time point
		taskTable[i].time -= startTime
		taskTable[i].task.Id = fmt.Sprintf("task%d", i)
		taskTable[i].task.LifeTime = time.Duration(float64(taskTable[i].task.LifeTime) * lifeRate)
	}

	// make the stream speed = tasknumPerSec (average)

	timeFactor := float64(len(taskTable)/tasknumCompressRate) / float64(taskTable[len(taskTable)-1].time/time.Second)

	log.Println(timeFactor)
	for i := range taskTable {
		taskTable[i].time = time.Duration(timeFactor * float64(taskTable[i].time))
	}

	//time bias
	minTime := time.Duration(timebias) * time.Second
	trucateIndex := 0
	for i := range taskTable {
		if taskTable[i].time > minTime {
			trucateIndex = i
			break
		}
	}
	taskTable = taskTable[trucateIndex:]

	startTime = taskTable[0].time
	for i := range taskTable {
		//adjust zero Time point
		taskTable[i].time -= startTime
		taskTable[i].task.Id = fmt.Sprintf("task%d", i)
	}

	var temp []SrcNode
	// don use the very large task
	for i := range taskTable {

		if taskTable[i].task.CpuRequest > maxResourceLimit {
			taskTable[i].task.CpuRequest = maxResourceLimit
			continue
		}

		if taskTable[i].task.MemoryRequest > maxResourceLimit {
			taskTable[i].task.MemoryRequest = maxResourceLimit
			continue
		}

		temp = append(temp, taskTable[i])
	}
	taskTable = temp
	for i := range taskTable {
		taskTable[i].task.Id = fmt.Sprintf("task%d", i)
	}

	newTable := applyLoadRate(taskTable, loadrate)

	var strTable [][]string
	for i := range newTable {
		var taskline []string = make([]string, 5)
		taskline[0] = fmt.Sprint(int64(newTable[i].time / time.Microsecond))
		taskline[1] = newTable[i].task.Id
		taskline[2] = fmt.Sprint(int64(newTable[i].task.LifeTime / time.Microsecond))
		taskline[3] = fmt.Sprint(newTable[i].task.CpuRequest)
		taskline[4] = fmt.Sprint(newTable[i].task.MemoryRequest)
		strTable = append(strTable, taskline)
	}

	top := []string{"time", "taskid", "tasklife", "taskCpu", "taskRam"}
	common.ListToCsv(strTable, top, outfile)

}

func applyLoadRate(taskTable []SrcNode, loadrate float64) []SrcNode {
	//loadrate
	var newTable []SrcNode
	//var loadrate = basicLoadRate * (float64(config.Val.NodeNum) / 1000.0)
	for i := range taskTable {
		for j := 0; j < int(loadrate); j++ {
			newtask := taskTable[i]
			newTable = append(newTable, newtask)
		}

		if rand.Float64() < loadrate-float64(int64(loadrate)) {
			newtask := taskTable[i]
			newTable = append(newTable, newtask)
		}
	}

	for i := range newTable {
		newTable[i].task.Id = fmt.Sprintf("task%d", i)
	}
	return newTable
}
