package core

// 请将添加的集群在这里注册
var ClusterMarket = map[string]func() Cluster{
	"Center":     BuildCenterCluster,
	"Dcss":       BuildDcssCluster,
	"ShareState": BuildShareStateCluster,
	// "Raft":       core.BuildRaftCluster,
	// "Sparrow":    core.BuildSparrowCluster,
}
