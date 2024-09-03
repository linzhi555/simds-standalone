package main

import (
	"log"
	"runtime"
	"simds-standalone/config"
	"simds-standalone/tracing/analyzer"
)

func main() {
	outputDir := config.Val.OutputDir
	taskLogFile := outputDir + "/" + config.Val.TaskEventsLogName
	netLogFile := outputDir + "/" + config.Val.NetEventsLogName

	log.Println("analyzing net events.....")
	analyzer.AnalyseNet(netLogFile, outputDir)
	log.Println("analyzing net events finished")

	runtime.GC()

	log.Println("analyzing task events.....")
	analyzer.AnalyseTasks(taskLogFile, outputDir)
	log.Println("analyzing task events finished")

}
