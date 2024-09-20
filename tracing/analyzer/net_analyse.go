package analyzer

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"simds-standalone/common"
	"sort"
	"strings"
	"time"
)

var _stage string

func _start(stage string) {
	_stage = stage
	log.Printf("	 %s...\n", _stage)
}
func _finish() {
	log.Printf("	 %s finished\n", _stage)
}

func AnalyseNet(logfile string, outdir string) {

	_start("loding message events")
	events := parseNetEventCSV(logfile)
	_finish()

	_start("analyze net message latency")
	//AnalyzeStageDuration(events, "send", "recv").RemoveFails().Output(outdir, "_netLantency")
	AnalyzeStageDuration(events, "send", "recv").RemoveFails().Output(outdir, "_netLatency")
	_finish()

	_start("analyze all cluster message rate")
	AnalyzeEventRate(events, "recv", 100).Output(outdir, "_allNet")
	_finish()

	_start("analyze busiesthost net events")
	busiestHost, busiestHostEvents := busiestHost(events)
	log.Println("     the most busiest host is ", hostTable[busiestHost])
	AnalyzeEventRate(busiestHostEvents, "recv", 100).Output(outdir, "_busiestHostNet")
	_finish()
}

var NET_EVENT_LOG_HEAD = []string{"time", "Id", "direction", "type", "from", "to"}

type NetEvent struct {
	Time   time.Time
	Id     common.UID
	IsSend bool
	Head   uint16
	From   uint16
	To     uint16
}

const (
	_NTime = iota
	_NId
	_NDirect
	_NType
	_NFrom
	_NTo
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

	err = common.IterateCsv(fs, nil, func(row []string) {
		_append(&res, row)
	})

	if err != nil {
		if strings.HasPrefix(err.Error(), "partial error:") {
			log.Println(err)
		} else {
			panic(err)
		}
	}

	sort.Sort(NetEventLine(res))
	reverseTable(hostCache, hostTable)
	reverseTable(headCache, headTable)

	return res
}

var hostCache = map[string]uint16{}
var hostTable = map[uint16]string{}
var hostNum uint16 = 0

func hostname2uint(name string) uint16 {
	i, ok := hostCache[name]
	if ok {
		return i
	} else {
		hostNum++
		hostCache[name] = hostNum
		return hostNum
	}
}

var headCache = map[string]uint16{}
var headTable = map[uint16]string{}
var HeadNum uint16 = 0

func reverseTable(from map[string]uint16, to map[uint16]string) {
	for k, v := range from {
		to[v] = k
	}
}

func head2uint(name string) uint16 {
	i, ok := headCache[name]
	if ok {
		return i
	} else {
		hostNum++
		hostCache[name] = hostNum
		return hostNum
	}
}

func _append(l *[]NetEvent, row []string) {

	// do use the self send message
	if row[_NFrom] == row[_NTo] {
		return
	}

	t, err := common.ParseTime(row[_NTime])
	if err != nil {
		panic(err)
	}

	*l = append(*l, NetEvent{
		Time: t,
		Id:   common.ReadUID(row[_NId]),
		IsSend: func(direct string) bool {
			if direct == "send" {
				return true
			} else if direct == "recv" {
				return false
			} else {
				panic("wrong direction")
			}
		}(row[_NDirect]),
		Head: head2uint(row[_NType]),
		From: hostname2uint(row[_NFrom]),
		To:   hostname2uint(row[_NTo]),
	})
}

func busiestHost(events NetEventLine) (uint16, NetEventLine) {
	var mostBusiestScore int = 0
	var mostBusiestIndex uint16 = 0
	var mostBusiestEvents NetEventLine

	for curIndex, name := range hostTable {
		if strings.HasPrefix(name, "simds-taskgen") {
			continue
		}

		var curEvents NetEventLine
		for _, e := range events {
			if e.From == curIndex || e.To == curIndex {
				curEvents = append(curEvents, e)
			}
		}

		curScore := AnalyzeEventRate(curEvents, "recv", 100).Highest()
		if curScore > mostBusiestScore {
			mostBusiestScore = curScore
			mostBusiestIndex = curIndex
			mostBusiestEvents = curEvents
		}
	}

	return mostBusiestIndex, mostBusiestEvents
}
