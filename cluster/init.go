package cluster

import (
	centrailzed "simds-standalone/cluster/centralized"
	"simds-standalone/cluster/dcss"
	"simds-standalone/cluster/sharestate"
	"simds-standalone/core"
)

// // 请将添加的集群在这里注册
var ClusterMarket = map[string]func() core.Cluster{
	"Center":     centrailzed.BuildCenterCluster,
	"Dcss":       dcss.BuildDcssCluster,
	"ShareState": sharestate.BuildShareStateCluster,
	// "Raft":       core.BuildRaftCluster,
	// "Sparrow":    core.BuildSparrowCluster,
}
