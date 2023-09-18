package main

import (
	"flag"
	"fmt"
	"strings"
)

var LogLevel = "Info"
var disableInfo = flag.Bool("q", false, "disable info")

func ParseAddr(addr string) (host, port string) {
	temp := strings.Split(addr, ":")
	if len(temp) != 2 {
		panic("addr is wrong")
	}
	return temp[0], temp[1]
}

func IsSameHost(addr0, addr1 string) bool {
	host1, _ := ParseAddr(addr0)
	host2, _ := ParseAddr(addr1)
	return host1 == host2
}

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
