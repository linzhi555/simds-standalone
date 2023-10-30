package main

import (
	"log"
	"time"

	"simds-standalone/common"
)

func main() {
	// 模拟性能分析,调试用
	common.StartPerf()
	defer common.StopPerf()

	// 创建所需集群
	var cluster Cluster
	if Config.Dcss {
		log.Println("run dcss cluster")
		cluster = BuildDCSSCluster()
	} else if Config.ShareState {
		log.Println("run share state cluster")
		cluster = BuildShareStateCluster()
	} else if Config.Center {
		log.Println("run centralized cluster")
		cluster = BuildCenterCluster()
	} else {
		panic("pleas specify which cluster to run")
	}

	LogConfig()
	time.Sleep(3 * time.Second)

	// 用ECS 运行该集群
	EcsRunCluster(cluster)

}
