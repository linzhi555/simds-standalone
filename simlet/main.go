package main

import (
	"log"
	"path"
	"time"

	"simds-standalone/cluster"
	"simds-standalone/cluster/base"
	"simds-standalone/common"
	"simds-standalone/config"
	"simds-standalone/simctl/k8s"
	"simds-standalone/simlet/svc"
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

func waitUitlClusterFullyStart(cluster *base.Cluster) *svc.RouterTable {
	cli, err := k8s.CreateReadonlyInContainerClient()
	if err != nil {
		panic(cli)
	}

	log.Println(cli.GetNamespace())
	pods := cli.GetPodsWithPrefix("simds-")
	log.Println(pods)

	lastNodesName := cluster.Nodes[len(cluster.Nodes)-1].GetHostName()
	for {
		err := cli.WaitUtilAllRunning([]string{lastNodesName})
		if err != nil {
			log.Println("waitting for ",lastNodesName," run:", err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	pods = cli.GetPodsWithPrefix("simds-")

	var rtable svc.RouterTable
	for _, pod := range pods {
		ip, _ := cli.GetPodIP(pod)
		log.Println("get pod ip:", pod, ip)
		rtable.Columns = append(rtable.Columns, &svc.AddrPair{ActorAddr: pod, SimletAddr: ip + ":8888"})
	}

	return &rtable
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

	routerTable := waitUitlClusterFullyStart(&cluster)
	simletServer := NewServerWithRouterTable(routerTable)

	// start serving
	go simletServer.RunInputServer()
	go simletServer.RunOutputThread()

	startSimletActor(initActor, simletServer)
	for {
	}

}
