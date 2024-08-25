package main

import (
	"log"
	"path"
	"simds-standalone/cluster"
	"simds-standalone/cluster/base"
	"simds-standalone/common"
	"simds-standalone/config"
	"time"
)

func startSimletActor(node base.Node, s *SimletServer) {
	os := NewActorOs(config.Val.NodeName)
	s.RegisterNewActor(os)
	node.SetOsApi(os)

	os.Send(base.Message{
		From:    node.GetHostName(),
		To:      node.GetHostName(),
		Content: "SignalBoot",
		Body:    base.Signal("SignalBoot"),
	})

	go func() {
		for {
			m := <-os.input
			node.Update(m)
		}
	}()
}

func main() {
	log.Println("simlet started as", config.Val.NodeName)

	// Init log file
	common.AppendLineCsvFile(NETWORK_EVENT_LOG_NAME, []string{"time", "type", "from", "to", "body"})
	common.AppendLineCsvFile(base.TASKS_EVENT_LOG_NAME, []string{"time", "taskid", "type", "nodeip", "cpu", "ram"})

	// base.InitLogs()
	config.LogConfig(path.Join(config.Val.OutputDir, "config.log"))

	// Initialize self as a specified node of cluster
	clusterBuilder, ok := cluster.ClusterMarket[config.Val.Cluster]
	if !ok {
		keys := make([]string, 0, len(cluster.ClusterMarket))
		for k := range cluster.ClusterMarket {
			keys = append(keys, k)
		}
		log.Panicln("wrong type of cluster,registed cluster is", keys)
	}

	var cluster base.Cluster = clusterBuilder()
	var initActor base.Node

	for _, n := range cluster.Nodes {
		if n.GetHostName() == config.Val.NodeName {
			initActor = n
			break
		}
	}

	simletServer := NewServer()

	// start serving
	go simletServer.RunInputServer()
	go simletServer.RunOutputThread()

	// only after the simlet server net is ready,then init the actor
	for !simletServer.inited.isTrue() {
		time.Sleep(time.Second)
	}

	startSimletActor(initActor, simletServer)
	for {
	}

}
