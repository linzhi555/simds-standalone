package main

import (
	"flag"
	"log"
	"path"
)

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
	info("output sorted events line...")
	events.Output(*outputDir)
	info("init cluster ...")
	c := InitCluster(events)
	info("analysing latency...")
	c.AnalyseSchedulerLatency(*outputDir)
	info("analysing task lifet time...")
	c.AnalyseTaskLifeTime(*outputDir)
	info("analysing cluster resource status curves...")
	curves := c.CalStatusCurves(*outputDir)

	info("output pngs ...")
	OutputAverageCpuRamCurve(path.Join(*outputDir, "clusterStatusAverage.png"), curves)
	OutputVarianceCpuRamCurve(path.Join(*outputDir, "clusterStatusVariance.png"), curves)
}
