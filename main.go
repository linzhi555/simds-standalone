package main

import (
	"log"
	"path"
	"runtime"
	"time"

	"simds-standalone/common"
	"simds-standalone/config"
	"simds-standalone/core"
)

func main() {

	runtime.GOMAXPROCS(int(config.Val.GoProcs))

	// 模拟性能分析,调试用
	common.StartPerf()
	defer common.StopPerf()

	core.InitLogs()
	config.LogConfig(path.Join(config.Val.OutputDir, "config.log"))

	// 请将添加的集群在这里注册
	clusterMarket := map[string]func() core.Cluster{
		"Dcss":       core.BuildDCSSCluster,
		"ShareState": core.BuildShareStateCluster,
		"Center":     core.BuildCenterCluster,
		"Raft":       core.BuildRaftCluster,
		"Sparrow":    core.BuildSparrowCluster,
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
	var cluster core.Cluster = clusterBuilder()
	time.Sleep(3 * time.Second)

	if config.Val.Debug {
		// 使用调度控制台模式运行集群
		core.EcsRunClusterDebug(cluster)
	} else {
		// 用ECS 运行该集群
		start := time.Now()
		core.EcsRunCluster(cluster)
		log.Println("simulation finished, time used: ", time.Since(start))
	}

}
