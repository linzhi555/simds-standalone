package main

import (
	"log"
	"path"
	"time"

	"simds-standalone/common"
)

func main() {
	// 模拟性能分析,调试用
	common.StartPerf()
	defer common.StopPerf()

	initLogs()
	LogConfig(path.Join(Config.OutputDir, "config.log"))

	clusterMarket := map[string]func() Cluster{
		"Dcss":       BuildDCSSCluster,
		"ShareState": BuildShareStateCluster,
		"Center":     BuildCenterCluster,
	}
	clusterBuilder, ok := clusterMarket[Config.Cluster]
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
