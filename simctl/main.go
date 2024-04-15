package main

import (
	"fmt"
	"log"
	"os/exec"
	"path"
	"simds-standalone/config"
	"simds-standalone/simctl/k8s"
)

func BuildImage() {
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
	cli.CreatePod("simds-node0", "simds-node0", config.Val.DockerImageRepo, []string{"sh", "-c", "/simlet --NodeName simds-node0  2>simlet_err.log; sleep 1000"})
	cli.CreateService("simds-node0-svc", "simds-node0", 8888, 32055)
	cli.CreatePod("simds-node1", "simds-node1", config.Val.DockerImageRepo, []string{"sh", "-c", "sleep 3;/simlet --NodeName simds-node1 2>simlet_err.log; sleep 1000"})
}

func main() {
	templatePath := config.Val.K8STemplatePath
	cli, err := k8s.ConnectToK8s(config.Val.K8SConfig, path.Join(templatePath, "pod_template.yaml"), path.Join(templatePath, "service_template.yaml"))
	if err != nil {
		panic(err)
	}
	fmt.Println(cli.GetPodsWithPrefix("simds"))
	BuildImage()
	clean(cli)
	//test(cli)

}
