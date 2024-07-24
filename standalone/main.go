package main

import (
	"fmt"
	"log"
	"path"
	"runtime"
	"time"

	"simds-standalone/cluster"
	"simds-standalone/common"
	"simds-standalone/config"
	"simds-standalone/core"
	"simds-standalone/standalone/engine"
)

func main() {

	runtime.GOMAXPROCS(int(config.Val.GoProcs))

	// 模拟性能分析,调试用
	common.StartPerf()
	defer common.StopPerf()

	// core.InitLogs()
	config.LogConfig(path.Join(config.Val.OutputDir, "config.log"))

	clusterBuilder, ok := cluster.ClusterMarket[config.Val.Cluster]
	if !ok {
		keys := make([]string, 0, len(cluster.ClusterMarket))
		for k := range cluster.ClusterMarket {
			keys = append(keys, k)
		}
		log.Panicln("wrong type of cluster,registed cluster is", keys)
	}

	// 创建所需集群
	var cluster core.Cluster = clusterBuilder()
	time.Sleep(3 * time.Second)

	simulator := engine.InitEngine(cluster)

	if config.Val.Debug {
		simulator.RunInConsole()
	} else {
		start := time.Now()
		simulator.Run()
		costTime := time.Since(start)
		simulator.Network.Os.LogInfo("costTime.log", fmt.Sprint(costTime))
	}

}
