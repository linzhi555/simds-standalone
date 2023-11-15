package main

import (
	"flag"
	"log"
	"path"
)

var taskLogFile = flag.String("taskLog", "./tasks_event.log", "the task event log files")
var netLogFile = flag.String("netLog", "./network_event.log", "the network log files")
var outputDir = flag.String("outputDir", "./target", "where to ouput the result")
var verbose = flag.Bool("verbose", false, "show the process information")
var nodeCPU = flag.Float64("CPU", 10.0, "the cpu of node")
var nodeMemory = flag.Float64("memory", 10.0, "the memory of node")

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
	taskevents := ReadTaskEventCsv(*taskLogFile)
	info("output sorted events line...")
	taskevents.Output(*outputDir)
	info("init cluster ...")
	c := InitCluster(taskevents)
	info("analysing latency...")
	c.AnalyseSchedulerLatency(*outputDir)
	info("analysing task lifet time...")
	c.AnalyseTaskLifeTime(*outputDir)
	info("analysing cluster resource status curves...")
	curves := c.CalStatusCurves(*outputDir)

	info("analysing net busy ...")
	AnalyseNet(*netLogFile, *outputDir)

	info("output pngs ...")
	OutputAverageCpuRamCurve(path.Join(*outputDir, "clusterStatusAverage.png"), curves)
	OutputVarianceCpuRamCurve(path.Join(*outputDir, "clusterStatusVariance.png"), curves)
}
