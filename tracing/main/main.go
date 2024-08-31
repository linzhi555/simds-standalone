package main

import (
	"simds-standalone/config"
	"simds-standalone/tracing/analyzer"
)

func main() {
	outputDir := config.Val.OutputDir
	taskLogFile := outputDir + "/" + config.Val.TaskEventsLogName
	netLogFile := outputDir + "/" + config.Val.NetEventsLogName

	analyzer.AnalyseTasks(taskLogFile, outputDir)
	analyzer.AnalyseNet(netLogFile, outputDir)
}
