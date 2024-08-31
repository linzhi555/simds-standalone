package analyzer

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"sort"
	"time"
)

func AnalyseNet(logfile string, outdir string) {
	events := parseNetEventCSV(logfile)
	eventsByHost := separateByReceiver(events)
	index := mostBusyHost(eventsByHost)

	AnalyzeEventRate(events, "recv", 100).Output(outdir, "allNet")
	AnalyzeEventRate(eventsByHost[index], "recv", 100).Output(outdir, "busiestHostNet")

	AnalyzeStageDuration(events, "send", "recv").Output(outdir, "netLantency")
}

var NET_EVENT_LOG_HEAD = []string{"Id", "time", "direction", "type", "from", "to"}

type NetEvent struct {
	Time   time.Time
	Id     string
	Direct string
	Type   string
	From   string
	To     string
}

const (
	_NTime = iota
	_Id
	_NDirect
	_NType
	_NFrom
	_NTo
)

type NetEventLine []*NetEvent

// for sort
func (l NetEventLine) Len() int      { return len(l) }
func (l NetEventLine) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l NetEventLine) Less(i, j int) bool {
	return l[i].Time.Before(l[j].Time)
}

// for  TaskLines Interface
func (l NetEventLine) GetID(i int) string            { return l[i].Id }
func (l NetEventLine) GetType(i int) string          { return l[i].Direct }
func (l NetEventLine) GetHappenTime(i int) time.Time { return l[i].Time }

func parseNetEventCSV(csvPath string) NetEventLine {
	fs, err := os.Open(csvPath)
	if err != nil {
		log.Fatal("can not open ", csvPath)
	}
	defer fs.Close()
	r := csv.NewReader(fs)
	var res []*NetEvent
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
			res = append(res, &NetEvent{
				Time:   t,
				Direct: row[_NDirect],
				Type:   row[_NType],
				From:   row[_NFrom],
				To:     row[_NTo],
			})

		}
	}

	sort.Sort(NetEventLine(res))

	return res
}

func mostBusyHost(events []NetEventLine) int {
	mostBusyIndex := 0
	mostBusyValue := -1
	for i, hostEvents := range events {
		busyValue := AnalyzeEventRate(hostEvents, "recv", 100).Highest()
		if busyValue > mostBusyValue {
			mostBusyValue = busyValue
			mostBusyIndex = i
		}
	}
	return mostBusyIndex
}

func separateByReceiver(allEvents []*NetEvent) []NetEventLine {
	table := map[string][]*NetEvent{}
	for _, event := range allEvents {
		table[event.To] = append(table[event.To], event)
	}

	res := []NetEventLine{}
	for _, list := range table {
		res = append(res, list)
	}

	return res
}
