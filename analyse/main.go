package main

import (
	"flag"
	"log"
	"path"

	"simds-standalone/common"
)

var timeRate = flag.Int("timeRate", 1, "the time rate of the simulation")
var logFile = flag.String("logFile", "./test.log", "the original log files")
var outputDir = flag.String("outputDir", "./target", "where to ouput the result")
var verbose = flag.Bool("verbose", false, "show the process information")

func init() {
	flag.Parse()
}

func info(s string) {
	if *verbose {
		log.Println(s)
	}
}

func main() {
	info("test start,reading csv and sort.....")
	events := ReadTaskEventCsv(*logFile)
	info("adjust time rate...")
	AdjustEventTimeByTimeRate(*timeRate, TaskEventLine(events))
	info("analysing latency...")
	events.AnalyseSchedulerLatency(*outputDir)
	info("analysing task lifet time...")
	events.AnalyseTaskLifeTime(*outputDir)

	info("output sorted events line...")
	outputlogfile := path.Join(*outputDir, "all_events.log")
	err := common.AppendLineCsvFile(outputlogfile, []string{"time", "taskid", "type", "nodeip", "cpu", "ram"})
	if err != nil {
		panic(err)
	}

	for _, event := range events {
		err = common.AppendLineCsvFile(outputlogfile, event.Strings())
		if err != nil {
			panic(err)
		}

	}

	info("analysing cluster status ...")
	c := InitCluster(events)
	curves := c.CalStatusCurves(events)

	info("output pngs ...")
	OutputAverageCpuRamCurve(path.Join(*outputDir, "clusterStatusAverage.png"), curves)
	OutputVarianceCpuRamCurve(path.Join(*outputDir, "clusterStatusVariance.png"), curves)
}
