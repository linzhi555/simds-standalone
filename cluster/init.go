package cluster

import (
	"simds-standalone/cluster/base"
	"simds-standalone/cluster/centralized"
	"simds-standalone/cluster/dcss"
	"simds-standalone/cluster/sharestate"
)

// // 请将添加的集群在这里注册
var ClusterMarket = map[string]func() base.Cluster{
	"Center":     centrailzed.BuildCenterCluster,
	"Dcss":       dcss.BuildDcssCluster,
	"ShareState": sharestate.BuildShareStateCluster,
	// "Raft":       core.BuildRaftCluster,
	// "Sparrow":    core.BuildSparrowCluster,
}
