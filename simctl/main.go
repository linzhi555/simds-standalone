package main

import (
	"fmt"
	"log"
	"os/exec"
	"path"
	"simds-standalone/config"
	"simds-standalone/core"
	"simds-standalone/simctl/k8s"
	"strings"
	"time"
)

func PushImage() {
	cmd := exec.Command("go", "build", "-o", "./target/simlet", "./simlet")
	if err := cmd.Run(); err != nil {
		log.Println("simlet Build Failed")
		log.Fatal(err)
	}
	log.Println("simlet Build Succssed")

	cmd = exec.Command("docker", "build", "-t", config.Val.DockerImageRepo, ".")
	if err := cmd.Run(); err != nil {
		log.Println("Image Build Failed")
		log.Fatal(err)
	}
	log.Println("Image Build Succssed")

	cmd = exec.Command("docker", "push", config.Val.DockerImageRepo)
	if err := cmd.Run(); err != nil {
		log.Println("Image Push Failed")
		log.Fatal(err)
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
		fmt.Println("deploy", node.GetHostName())
		if strings.HasPrefix(node.GetHostName(), "taskgen") {
			time.Sleep(time.Second * 10)
		}

		name := fmt.Sprintf("simds-%s", node.GetHostName())
		cli.CreatePod(name, name, config.Val.DockerImageRepo, []string{"sh",
			"-c",
			fmt.Sprintf("/simlet --Cluster %s --NodeName %s  >simlet.log 2>simlet_err.log; sleep 20000", config.Val.Cluster, node.GetHostName())})
		cli.CreateService(fmt.Sprintf("%s-svc", name), name, 8888, 32055+i)
	}

}

func main() {
	templatePath := config.Val.K8STemplatePath
	cli, err := k8s.ConnectToK8s(config.Val.K8SConfig, path.Join(templatePath, "pod_template.yaml"), path.Join(templatePath, "service_template.yaml"))
	if err != nil {
		panic(err)
	}
	fmt.Println(cli.GetPodsWithPrefix("simds"))
	PushImage()
	clean(cli)
	test(cli)

}
