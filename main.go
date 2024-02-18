package main

import (
	"log"
	"path"
	"simds-standalone/common"
	"simds-standalone/config"
	"time"
)

func main() {
	// 模拟性能分析,调试用
	common.StartPerf()
	defer common.StopPerf()

	initLogs()
	config.LogConfig(path.Join(config.Val.OutputDir, "config.log"))

	// 请将添加的集群在这里注册
	clusterMarket := map[string]func() Cluster{
		"Dcss":       BuildDCSSCluster,
		"ShareState": BuildShareStateCluster,
		"Center":     BuildCenterCluster,
		"Raft":       BuildRaftCluster,
		"Sparrow":    BuildSparrowCluster,
	}
	clusterBuilder, ok := clusterMarket[config.Val.Cluster]
	if !ok {
		keys := make([]string, 0, len(clusterMarket))
		for k := range clusterMarket {
			keys = append(keys, k)
		}
		log.Panicln("wrong type of cluster,registed cluster is", keys)
	}

	// 创建所需集群
	var cluster Cluster = clusterBuilder()
	time.Sleep(3 * time.Second)

	// 用ECS 运行该集群
	EcsRunCluster(cluster)

}
