package main

import (
	"fmt"
	"log"
	"path"
	"runtime"
	"time"

	"simds-standalone/cluster"
	"simds-standalone/cluster/base"
	"simds-standalone/common"
	"simds-standalone/config"
	"simds-standalone/standalone/engine"
)

func main() {

	runtime.GOMAXPROCS(int(config.Val.GoProcs))

	// 模拟性能分析,调试用
	common.StartPerf()
	defer common.StopPerf()

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
	var cluster base.Cluster = clusterBuilder()
	time.Sleep(3 * time.Second)

	simulator := engine.InitEngine(cluster)

	if config.Val.Debug {
		//simulator.RunInConsole()
		simulator.GuiDebugging()
	} else {
		start := time.Now()
		simulator.Run()
		costTime := time.Since(start)

		common.AppendLineCsvFile(
			path.Join(config.Val.OutputDir, "costTime.log"),
			[]string{fmt.Sprint(costTime)},
		)
	}

}
