package main

import (
	"log"
	"path"
	"time"

	"simds-standalone/cluster"
	"simds-standalone/cluster/base"
	"simds-standalone/cluster/lib"
	"simds-standalone/config"
	"simds-standalone/simctl/k8s"
	"simds-standalone/simlet/svc"
	"simds-standalone/tracing/rules"
)

func startSimletActor(actor base.Actor, s *SimletServer) {
	os := NewActorOs(config.Val.NodeName)
	s.RegisterNewActor(os)
	actor.SetOsApi(os)

	boolSignal := lib.Signal("SignalBoot")
	os.Send(base.Message{
		From: actor.GetHostName(),
		To:   actor.GetHostName(),
		Head: string(boolSignal),
		Body: boolSignal,
	})

	go func() {
		for {
			msg := <-os.input
			rules.CheckRulesThenExec(rules.MsgDealRules, time.Now(), &msg)
			actor.Update(msg)
			rules.CheckRulesThenExec(rules.MsgFinishRules, time.Now(), &msg)
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

	lastNodesName := cluster.Nodes[len(cluster.Nodes)-1].Actors[0].GetHostName()
	for {
		err := cli.WaitUtilAllRunning([]string{lastNodesName})
		if err != nil {
			log.Println("waitting for ", lastNodesName, " run:", err)
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

	// write some log file first
	config.LogConfig(path.Join(config.Val.OutputDir, "config.log"))
	rules.InitTracing()

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
	var initActor base.Actor

	for _, n := range cluster.Nodes {
		if n.Actors[0].GetHostName() == config.Val.NodeName {
			initActor = n.Actors[0]
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
		time.Sleep(time.Hour)
	}

}
