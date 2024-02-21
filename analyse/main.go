package main

import (
	"fmt"
	"log"
	"path"
	"simds-standalone/config"

	"github.com/spf13/pflag"
)

//var taskLogFile = pflag.String("taskLog", "./tasks_event.log", "the task event log files")
//var netLogFile = pflag.String("netLog", "./network_event.log", "the network log files")
//var outputDir = pflag.String("outputDir", "./target", "where to ouput the result")
//var verbose = pflag.Bool("verbose", false, "show the process information")

func init() {
	pflag.Parse()
}

func info(s string) {
	log.Println(s)
}

func main() {
	outputDir := config.Val.OutputDir
	taskLogFile := outputDir + "/" + "tasks_event.log"
	netLogFile := outputDir + "/" + "network_event.log"

	info("test start,reading csv and sort.....")
	fmt.Println(taskLogFile)
	taskevents := ReadTaskEventCsv(taskLogFile)
	info("output sorted events line...")
	taskevents.Output(outputDir)
	info("init cluster ...")
	c := InitCluster(taskevents)
	info("analysing latency...")
	c.AnalyseSchedulerLatency(outputDir)
	info("analysing task lifet time...")
	c.AnalyseTaskLifeTime(outputDir)
	info("analysing cluster resource status curves...")
	curves := c.CalStatusCurves(outputDir)

	info("analysing net busy ...")
	AnalyseNet(netLogFile, outputDir)

	info("output pngs ...")
	OutputAverageCpuRamCurve(path.Join(outputDir, "clusterStatusAverage.png"), curves)
	OutputVarianceCpuRamCurve(path.Join(outputDir, "clusterStatusVariance.png"), curves)
}
