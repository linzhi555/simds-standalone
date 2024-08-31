package analyzer

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"simds-standalone/common"
	"sort"
	"time"
)

var NET_EVENT_LOG_HEAD = []string{"time", "direction", "type", "from", "to"}

// ooutput
const (
	AllCurveLog      = "all_net_curve.log"
	MostBusyCurveLog = "most_busy_curve.log"
)

type NetEvent struct {
	Time   time.Duration
	Direct string
	Type   string
	From   string
	To     string
}

const (
	_NTime = iota
	_NDirect
	_NType
	_NFrom
	_NTo
)

type NetEventLine []*NetEvent

func (l NetEventLine) Len() int      { return len(l) }
func (l NetEventLine) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l NetEventLine) Less(i, j int) bool {
	return l[i].Time < l[j].Time
}

type NetSpeedDot struct {
	TimeMs int
	Amount int
}

func OutputNetSpeedCurve(outputFile string, curve []NetSpeedDot) {
	if common.IsFileExist(outputFile) {
		err := os.Remove(outputFile)
		if err != nil {
			panic(err)
		}
	}
	err := common.AppendLineCsvFile(outputFile, []string{"time_ms", "amount"})
	if err != nil {
		panic(err)
	}

	for _, event := range curve {
		err = common.AppendLineCsvFile(outputFile, []string{fmt.Sprint(event.TimeMs), fmt.Sprint(event.Amount)})
		if err != nil {
			panic(err)
		}
	}

}

func parseNetEventCSV(csvPath string) []*NetEvent {
	fs, err := os.Open(csvPath)
	if err != nil {
		log.Fatal("can not open ", csvPath)
	}
	defer fs.Close()
	r := csv.NewReader(fs)
	var res []*NetEvent
	var startTime time.Time
	for i := 0; ; i++ {
		row, err := r.Read()
		if err != nil && err != io.EOF {
			panic("fail to read" + err.Error())
		}
		if err == io.EOF {
			break
		}
		if i == 0 {
			continue
		} else {
			t, err := time.Parse(time.RFC3339Nano, row[_NTime])
			if err != nil {
				panic(err)
			}
			if i == 1 {
				startTime = t
			}
			res = append(res, &NetEvent{
				Time:   t.Sub(startTime),
				Direct: row[_NDirect],
				Type:   row[_NType],
				From:   row[_NFrom],
				To:     row[_NTo],
			})

		}
	}

	sort.Sort(NetEventLine(res))

	bias := res[0].Time
	for i := range res {
		res[i].Time += -1 * bias
	}

	return res
}

func mostBusyHost(events [][]*NetEvent) int {
	mostBusyIndex := 0
	mostBusyValue := -1
	for i, hostEvents := range events {
		busyValue := highestPeek(hostEvents)
		if busyValue > mostBusyValue {
			mostBusyValue = busyValue
			mostBusyIndex = i
		}
	}
	return mostBusyIndex
}

func separateByReceiver(allEvents []*NetEvent) [][]*NetEvent {
	table := map[string][]*NetEvent{}
	for _, event := range allEvents {
		table[event.To] = append(table[event.To], event)
	}

	res := [][]*NetEvent{}
	for _, list := range table {
		res = append(res, list)
	}

	return res
}

func netSpeedCurveAnalyse(events []*NetEvent) []NetSpeedDot {
	lastTime := events[len(events)-1].Time
	var result []NetSpeedDot = make([]NetSpeedDot, lastTime/(100*time.Millisecond)+1)
	for i := range result {
		result[i].TimeMs = i * 100
		result[i].Amount = 0
	}

	for _, e := range events {
		nth100Ms := int64(e.Time / (100 * time.Millisecond))
		result[nth100Ms].Amount += 1
	}
	return result

}

// return the highes peek value of network in whole process
func highestPeek(events []*NetEvent) int {
	// records is the map record net message number of n th 10ms key: n th 10ms,value : message number
	records := map[int64]int{}

	for _, e := range events {
		nthMs := int64(e.Time / (100 * time.Millisecond))
		if _, ok := records[nthMs]; ok {
			records[nthMs] += 1
		} else {
			records[nthMs] = 1
		}
	}

	max := func(list map[int64]int) int {
		maxValue := 0
		for _, value := range list {
			if value > maxValue {
				maxValue = value
			}
		}
		return maxValue
	}

	return max(records)
}

func AnalyseNet(csvPath string, outputDir string) {
	allEvents := parseNetEventCSV(csvPath)
	allEventsByHost := separateByReceiver(allEvents)
	index := mostBusyHost(allEventsByHost)

	allNetSpeedCurve := netSpeedCurveAnalyse(allEvents)
	mostBusyNetSpeedCurve := netSpeedCurveAnalyse(allEventsByHost[index])
	OutputNetSpeedCurve(path.Join(outputDir, AllCurveLog), allNetSpeedCurve)
	OutputNetSpeedCurve(path.Join(outputDir, MostBusyCurveLog), mostBusyNetSpeedCurve)
}
