package main

import (
	"fmt"
	"log"
	"simds-standalone/config"

	"github.com/spf13/pflag"
)

func init() {
	pflag.Parse()
}

func info(s string) {
	log.Println(s)
}

func main() {
	outputDir := config.Val.OutputDir
	taskLogFile := outputDir + "/" + config.Val.TaskEventsLogName
	netLogFile := outputDir + "/" + config.Val.NetEventsLogName

	info("test start,reading csv and sort.....")
	fmt.Println(taskLogFile)
	taskevents := ReadTaskEventCsv(taskLogFile)
	info("output sorted events line...")
	taskevents.Output(outputDir)
	info("output task submit events")
	taskevents.OutputTaskSubmitRate(outputDir)
	info("init cluster ...")
	c := InitCluster(taskevents)
	info("analysing latency...")
	c.AnalyseSchedulerLatency(outputDir)
	info("analysing task lifet time...")
	c.AnalyseTaskLifeTime(outputDir)
	info("analysing cluster resource status curves...")
	c.CalStatusCurves(outputDir)

	info("analysing net busy ...")
	AnalyseNet(netLogFile, outputDir)

	info("output pngs ...")
}
