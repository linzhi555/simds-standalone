package main

import (
	"flag"
	"fmt"
)

var LogLevel = "Info"
var disableInfo = flag.Bool("q", false, "disable info")

func init() {
	if *disableInfo {
		LogLevel = "Error"
	}

}

func LogInfo(ecs *ECS, entity EntityName, ins ...interface{}) {
	if LogLevel != "Info" {
		return
	}
	fmt.Print(GetEntityTime(ecs, entity), " ", "Info", " ", entity, " ")
	for _, item := range ins {
		fmt.Print(item, " ")
	}
	fmt.Println()
}
