package main

import "fmt"

func LogInfo(ecs *ECS, entity EntityName, ins ...interface{}) {
	fmt.Print(GetEntityTime(ecs, entity), " ", "Info", " ", entity, " ")
	for _, item := range ins {
		fmt.Print(item, " ")
	}
	fmt.Println()
}
