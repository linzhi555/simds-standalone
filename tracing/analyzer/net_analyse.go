package analyzer

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"simds-standalone/common"
	"sort"
	"time"
)

func AnalyseNet(logfile string, outdir string) {

	log.Println("	 loding message events...")
	events := parseNetEventCSV(logfile)
	log.Println("	 loading message events finished")

	log.Println("	 analyze all cluster message rate...")
	AnalyzeEventRate(events, "recv", 100).Output(outdir, "allNet")
	log.Println("	 analyze all cluster message rate finished")

	log.Println("	 analyze net message latency...")
	AnalyzeStageDuration(events, "send", "recv").Output(outdir, "netLantency")
	log.Println("	 analyze net message latency finished")

	//host := mostBusyHost(events)
	//AnalyzeEventRate(eventsByHost[index], "recv", 100).Output(outdir, "busiestHostNet")
}

var NET_EVENT_LOG_HEAD = []string{"time", "Id", "direction", "type", "from", "to"}

type NetEvent struct {
	Time   time.Time
	Id     common.UID
	IsSend bool
	//Head   string
	//From   string
	//To     string
}

const (
	_NTime = iota
	_NId
	_NDirect
	//_NType
	//_NFrom
	//_NTo
)

type NetEventLine []NetEvent

// for sort
func (l NetEventLine) Len() int      { return len(l) }
func (l NetEventLine) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l NetEventLine) Less(i, j int) bool {
	return l[i].Time.Before(l[j].Time)
}

// for  TaskLines Interface
func (l NetEventLine) GetID(i int) string { return fmt.Sprint(string(l[i].Id[:])) }
func (l NetEventLine) GetType(i int) string {
	if l[i].IsSend {
		return "send"
	} else {
		return "recv"
	}

}
func (l NetEventLine) GetHappenTime(i int) time.Time { return l[i].Time }

func parseNetEventCSV(csvPath string) NetEventLine {
	linenum, err := common.CountLines(csvPath)
	if err != nil {
		panic("linum count fail: " + err.Error())
	}

	log.Println("		prepare reading csv  of lines:", linenum)

	runtime.GC()
	var res []NetEvent = make([]NetEvent, 0, linenum-1)

	fs, err := os.Open(csvPath)
	if err != nil {
		log.Fatal("can not open ", csvPath)
	}
	defer fs.Close()
	r := csv.NewReader(fs)
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
			t, err := common.ParseTime(row[_NTime])
			if err != nil {
				panic(err)
			}
			res = append(res, NetEvent{
				Time: t,
				Id:   common.ReadUID(row[_NTime]),
				IsSend: func(direct string) bool {
					if direct == "send" {
						return true
					} else if direct == "recv" {
						return false
					} else {
						panic("wrong direction")
					}
				}(row[_NDirect]),
				//Head:   row[_NType],
				//From:   row[_NFrom],
				//To:     row[_NTo],
			})

		}
	}
	sort.Sort(NetEventLine(res))
	return res
}

func mostBusyHost(events NetEventLine) string {
	var res string

	//for i := range events{

	//}

	return res
}
