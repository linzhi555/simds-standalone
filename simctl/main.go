package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path"
	"time"

	"simds-standalone/cluster"
	"simds-standalone/cluster/base"
	"simds-standalone/common"
	"simds-standalone/config"
	"simds-standalone/simctl/k8s"
)

func PushImage() {
	cmd := exec.Command("go", "build", "-o", "./target/simlet", "./simlet")
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Println("simlet Build Failed")
		log.Fatal(string(output))
	}
	log.Println("simlet Build Succssed")

	cmd = exec.Command("docker", "build", "--build-arg", fmt.Sprintf("Config=%s", config.Val.ConfigPath), "-t", config.Val.PushImageRepo, ".")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Image Build Failed")
		log.Fatal(string(output))
	}
	log.Println("Image Build Succssed")

	cmd = exec.Command("docker", "push", config.Val.PushImageRepo)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Println("Image Push Failed")
		log.Fatal(string(output))
	}
	log.Println("Image Push Succssed")
}

func Clean(cli *k8s.K8sClient) {
	cli.DeletePodsWithPrefix("simds")
	cli.DeleteServiceWithPrefix("simds")
}

func Deploy(cli *k8s.K8sClient) {
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

	// create pod
	for _, node := range cluster.Nodes {
		name := node.Actors[0].GetHostName()
		fmt.Println("deploy", name)

		cli.CreatePod(name, name, config.Val.PullImageRepo,
			[]string{"sh",
				"-c",
				fmt.Sprintf(
					"/simlet --Cluster %s --NodeName %s  >simlet.log 2>simlet_err.log; sleep 20000",
					config.Val.Cluster, name,
				),
			},
		)

		cli.WaitUtilAllRunning([]string{name})
		ip, err := cli.GetPodIP(name)
		if err != nil {
			panic(err)
		}
		podAddr := ip + ":8888"
		log.Println(name, podAddr)
	}
}

const threadNum = 1
func CollectResult(cli *k8s.K8sClient) {
	mergeCsvOfMultiplePods(cli, cli.GetPodsWithPrefix("simds-taskgen"), "tasks_event.log", path.Join(config.Val.OutputDir, "tasks_event.log"))
	mergeCsvOfMultiplePods(cli, cli.GetPodsWithPrefix("simds"), "network_event.log", path.Join(config.Val.OutputDir, "network_event.log"))
}

func mergeCsvOfMultiplePods(cli *k8s.K8sClient, pods []string, logfile string, outfile string) {
	var num = len(pods)
	var bufferCh = make(chan *bytes.Buffer, num)

	for i := 0; i < threadNum; i++ {
		go func(threadId int) {
			for j := 0; j < num; j++ {
				if j%threadNum == threadId {
					var b bytes.Buffer
					err := cli.Download(pods[j], "c1", logfile, &b)
					if err != nil {
						log.Println(err)
						bufferCh <- nil
					} else {
						bufferCh <- &b
					}
				}
			}
		}(i)
	}

	// merge csv

	var AllTable [][]string
	var TableTop []string

	for i := 0; i < num; i++ {
		log.Println("parallel down load file to ", outfile, " ", i, "/", num)
		b := <-bufferCh
		if b != nil {
			table, top := common.BytesCsvToList(b)
			AllTable = append(AllTable, table...)
			if TableTop == nil {
				TableTop = top
			}
		}
	}
	common.ListToCsv(AllTable, TableTop, outfile)
}

func main() {

	templatePath := config.Val.K8STemplatePath
	cli, err := k8s.ConnectToK8s(
		config.Val.K8SConfig,
		path.Join(templatePath, "pod_template.yaml"),
		path.Join(templatePath, "service_template.yaml"),
	)

	if err != nil {
		panic(err)
	}

	if config.Val.CleanMode {
		Clean(cli)
		return
	}

	fmt.Println(cli.GetPodsWithPrefix("simds"))
	Clean(cli)
	PushImage()
	Deploy(cli)
	time.Sleep(2 * time.Duration(config.Val.SimulateDuration) * time.Millisecond)
	CollectResult(cli)
}
