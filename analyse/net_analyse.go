package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"simds-standalone/common"
)

type NetEvent struct {
	TimeMS int64
	Type   string
	From   string
	To     string
}

func OutputNetEventList(outputDir string, events []NetEvent) {
	outputlogfile := path.Join(outputDir, "network_most_busy.log")
	err := os.Remove(outputlogfile)
	if err != nil {
		log.Println(err)
	}

	err = common.AppendLineCsvFile(outputlogfile, []string{"time_ms", "type", "from", "to"})
	if err != nil {
		panic(err)
	}

	for _, event := range events {
		err = common.AppendLineCsvFile(outputlogfile, []string{fmt.Sprint(event.TimeMS), event.Type, event.From, event.To})
		if err != nil {
			panic(err)
		}
	}

}

func parseNetEventCSV(table [][]string) []NetEvent {
	var res []NetEvent
	for _, l := range table {
		res = append(res, NetEvent{
			TimeMS: common.Str_to_int64(l[0]),
			Type:   l[1],
			From:   l[2],
			To:     l[3],
		})
	}
	return res
}

func mostBusyHost(events [][]NetEvent) int {
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

func separateByReceiver(allEvents []NetEvent) [][]NetEvent {
	table := map[string][]NetEvent{}
	for _, event := range allEvents {
		table[event.To] = append(table[event.To], event)
	}

	res := [][]NetEvent{}
	for _, list := range table {
		res = append(res, list)
	}

	return res
}

// return the highes peek value of network in whole process
func highestPeek(events []NetEvent) int {
	// records is the map record net message number of n th 10ms key: n th 10ms,value : message number
	records := map[int64]int{}

	for _, e := range events {
		nthMs := e.TimeMS / 10
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
	table, _ := common.CsvToList(csvPath)
	allEvents := parseNetEventCSV(table)
	allEventsByHost := separateByReceiver(allEvents)
	index := mostBusyHost(allEventsByHost)
	OutputNetEventList(outputDir, allEventsByHost[index])

}
