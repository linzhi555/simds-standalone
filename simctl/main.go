package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path"
	"simds-standalone/common"
	"simds-standalone/config"
	"simds-standalone/core"
	"simds-standalone/simctl/k8s"
	"strings"
	"time"
)

func PushImage() {
	cmd := exec.Command("bash","-c","CGO_ENABLED=0" ,"go", "build", "-o", "./target/simlet", "./simlet")
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Println("simlet Build Failed")
		log.Fatal(string(output))
	}
	log.Println("simlet Build Succssed")

	cmd = exec.Command("docker", "build", "--build-arg", fmt.Sprintf("Config=%s", config.Val.ConfigPath), "-t", config.Val.DockerImageRepo, ".")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Image Build Failed")
		log.Fatal(string(output))
	}
	log.Println("Image Build Succssed")

	cmd = exec.Command("docker", "push", config.Val.DockerImageRepo)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Println("Image Push Failed")
		log.Fatal(string(output))
	}
	log.Println("Image Push Succssed")
}

func clean(cli *k8s.K8sClient) {
	cli.DeletePodsWithPrefix("simds")
	cli.DeleteServiceWithPrefix("simds")
}
func test(cli *k8s.K8sClient) {
	// Initialize self as a specified node of cluster
	clusterBuilder, ok := core.ClusterMarket[config.Val.Cluster]
	if !ok {
		keys := make([]string, 0, len(core.ClusterMarket))
		for k := range core.ClusterMarket {
			keys = append(keys, k)
		}
		log.Panicln("wrong type of cluster,registed cluster is", keys)
	}

	var cluster core.Cluster = clusterBuilder()
	for i, node := range cluster.Nodes {
		if strings.HasPrefix(node.GetHostName(), "taskgen") {
			time.Sleep(time.Second * 20)
		}
		if strings.HasPrefix(node.GetHostName(), "storage") {
			time.Sleep(time.Second * 20)
		}
		fmt.Println("deploy", node.GetHostName())
		name := fmt.Sprintf("simds-%s", node.GetHostName())
		cli.CreatePod(name, name, config.Val.DockerImageRepo, []string{"sh",
			"-c",
			fmt.Sprintf(
				"tc qdisc add dev eth0 root netem delay %dus %dus; /simlet --Cluster %s --NodeName %s  >simlet.log 2>simlet_err.log; sleep 20000",
				config.Val.NetLatency*1000, int32(float32(config.Val.NetLatency*1000)*0.15), config.Val.Cluster, node.GetHostName())})
		cli.CreateService(fmt.Sprintf("%s-svc", name), name, 8888, 30100+i)
	}
}

func collectResult(cli *k8s.K8sClient) {
	mergeCsvOfMultiplePods(cli, cli.GetPodsWithPrefix("simds-taskgen"), "tasks_event.log", path.Join(config.Val.OutputDir, "tasks_event.log"))
	mergeCsvOfMultiplePods(cli, cli.GetPodsWithPrefix("simds"), "network_event.log", path.Join(config.Val.OutputDir, "network_event.log"))
}

func mergeCsvOfMultiplePods(cli *k8s.K8sClient, pods []string, logfile string, outfile string) {
	var num = len(pods)
	var bufferCh = make(chan *bytes.Buffer, num)

	// use 4 thread to download file
	threadNum := 4
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
	cli, err := k8s.ConnectToK8s(config.Val.K8SConfig, path.Join(templatePath, "pod_template.yaml"), path.Join(templatePath, "service_template.yaml"))
	if err != nil {
		panic(err)
	}

	if config.Val.CleanMode {
		clean(cli)
		return
	}

	fmt.Println(cli.GetPodsWithPrefix("simds"))
	clean(cli)
	PushImage()
	test(cli)
	time.Sleep(100 * time.Second)
	collectResult(cli)
}
