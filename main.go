package main

import (
	"flag"
	"simds-standalone/common"
)

var dcss = flag.Bool("dcss", false, "run dcss")
var shareState = flag.Bool("share", false, "run share state cluster")

func init() {
	flag.Parse()
}

func main() {
	// 模拟性能分析,调试用
	common.StartPerf()
	defer common.StopPerf()

	// 创建所需集群
	var cluster Cluster
	if *dcss {
		cluster = BuildDCSSCluster()
	} else if *shareState {
		cluster = BuildShareStateCluster()

	} else {
		cluster = BuildCenterCluster()
	}

	// 用ECS 运行该集群
	EcsRunCluster(cluster)

}
