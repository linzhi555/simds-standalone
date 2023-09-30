package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"time"
)

var LogLevel = "Info"
var disableInfo = flag.Bool("q", false, "disable info")

func ParseAddr(addr string) (host, port string) {
	temp := strings.Split(addr, ":")
	if len(temp) != 2 {
		panic("addr is wrong:"+addr)
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

var startTime = time.Now()

func init() {
	f, err := os.OpenFile("./test.log", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	f.WriteString("time,taskid,type,nodeip,cpu,ram\n")
	f.Close()

}

func TaskEventLog(t int32, task *TaskInfo, host EntityName) {
	timestr := fmt.Sprint(startTime.Add(time.Duration(t) * time.Millisecond).Format(time.RFC3339Nano))
	AppendLineCsvFile("./test.log", []string{timestr, task.Id, task.Status, string(host), fmt.Sprint(task.CpuRequest), fmt.Sprint(task.MemoryRequest)})
}
func AssertTypeIsNotPointer(v interface{}) {
	typestr := fmt.Sprint(reflect.TypeOf(v))
	if strings.HasPrefix(typestr, "*") {
		panic(typestr + " is ponter type")
	}
}

func shuffleStringSlice(slice []string) {
	rand.Shuffle(len(slice), func(i, j int) { slice[i], slice[j] = slice[j], slice[i] })
}
