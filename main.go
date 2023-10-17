package main

import (
	"simds-standalone/common"
	"time"
)

func main() {
	// 模拟性能分析,调试用
	common.StartPerf()
	defer common.StopPerf()

	LogConfig()
	time.Sleep(3 * time.Second)

	// 创建所需集群
	var cluster Cluster
	if Config.Dcss {
		cluster = BuildDCSSCluster()
	} else if Config.ShareState {
		cluster = BuildShareStateCluster()

	} else {
		cluster = BuildCenterCluster()
	}

	// 用ECS 运行该集群
	EcsRunCluster(cluster)

}
