package main

import (
	"fmt"
	"time"
)

// 创造一个Raft集群
func BuildRaftCluster() Cluster {
	var cluster = createCluster()

	var nodes []Node
	for i := 0; i < int(4); i++ {
		nodeName := fmt.Sprintf("node%d", i)
		nodes = append(nodes, Node{
			nodeName,
			[]NodeComponent{
				NewRaftManager(nodeName),
			},
		})
	}
	cluster.Nodes = nodes
	cluster.RegisterFunc(CRaftManager, RaftSetup, RaftUpdate)

	return cluster
}

func RaftSetup(c interface{}) {
	raft := c.(*RaftManager)
	raft.Role = Follower
	LogInfo(raft.Os,raft.Os.Net().GetAddr(), "started")

}

func RaftUpdate(c interface{}) {
	raft := c.(*RaftManager)

	SinceLastHeartBeat := raft.Os.GetTime().Sub(raft.LastHeartBeat)

	if SinceLastHeartBeat > time.Second {
		raft.LastHeartBeat = raft.Os.GetTime()
		LogInfo(raft.Os, SinceLastHeartBeat)
	}

}
