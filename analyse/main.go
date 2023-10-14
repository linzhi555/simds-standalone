package main

import (
	"flag"
	"log"
	"path"

	"simds-standalone/common"
)

var timeRate = flag.Int("timeRate", 1, "the time rate of the simulation")
var logFile = flag.String("logFile", "./test.log", "the original log files")
var taskStartFlag = flag.String("startFlag", "start", "the start event type,used in calculating the latency,can be changed to other type,to calculate different latency")
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
	events := ReadTaskEventCsv("./test.log")
	info("adjust time rate...")
	AdjustEventTimeByTimeRate(*timeRate, TaskEventLine(events))
	info("analysing latency...")
	events.AnalyseSchedulerLatency(*outputDir)
	info("analysing task lifet time...")
	events.AnalyseTaskLifeTime(*outputDir)

	info("output sorted events line...")
	outputlogfile := path.Join(*outputDir, "all_events.log")
	common.AppendLineCsvFile(outputlogfile, []string{"time", "taskid", "type", "nodeip", "cpu", "ram"})
	for _, event := range events {
		common.AppendLineCsvFile(outputlogfile, event.Strings())
	}

	info("analysing cluster status ...")
	c := InitCluster(events)
	curves := c.CalStatusCurves(events)

	info("output pngs ...")
	OutputAverageCpuRamCurve(path.Join(*outputDir, "clusterStatusAverage.png"), curves)
	OutputVarianceCpuRamCurve(path.Join(*outputDir, "clusterStatusVariance.png"), curves)
}
