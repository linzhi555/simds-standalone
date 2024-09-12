package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"time"

	"simds-standalone/cluster"
	"simds-standalone/cluster/base"
	"simds-standalone/config"
	"simds-standalone/simctl/k8s"
)

func _run(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalln(cmd.Args, "Failed")
	} else {
		log.Println(cmd.Args, "Succssed")
	}
}

func PushImage() {
	_run(
		exec.Command("go", "build", "-o", "./target/simlet", "./simlet"),
	)

	_run(
		exec.Command("docker", "build", "--build-arg", fmt.Sprintf("Config=%s", config.Val.ConfigPath), "-t", config.Val.PushImageRepo, "."),
	)

	_run(
		exec.Command("docker", "push", config.Val.PushImageRepo),
	)
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
		name := node.Actors[0].GetAddress()
		fmt.Println("deploy", name)

		cli.CreatePod(name, name, config.Val.PullImageRepo,
			[]string{"sh",
				"-c",
				fmt.Sprintf(
					"/simlet --Cluster %s --NodeName %s  >simlet.log 2>simlet_err.log; sleep 36000",
					config.Val.Cluster, name,
				),
			},
		)

		err := cli.WaitUtilAllRunning([]string{name})
		if err != nil {
			panic(err)
		}
		ip, err := cli.GetPodIP(name)
		if err != nil {
			panic(err)
		}
		podAddr := ip + ":8888"
		log.Println(name, podAddr)
	}
}

func CollectResult(cli *k8s.K8sClient) {
	// 下载任务日志文件，只存在在taskgen类型的pod中
	collectAndMerge(cli,
		cli.GetPodsWithPrefix("simds-taskgen"),
		config.Val.TaskEventsLogName,
		path.Join(config.Val.OutputDir, config.Val.TaskEventsLogName),
	)

	// 下载网络运输日志文件，在所有类型的节点中
	collectAndMerge(cli,
		cli.GetPodsWithPrefix("simds"),
		config.Val.NetEventsLogName,
		path.Join(config.Val.OutputDir, config.Val.NetEventsLogName),
	)
}

const downloadThreads = 1

// 下载所有pod的csv文件,汇集到一个文件中
func collectAndMerge(cli *k8s.K8sClient, pods []string, logfile string, outfile string) {
	var num = len(pods)
	var bufferCh = make(chan *bytes.Buffer, num)

	for i := 0; i < downloadThreads; i++ {
		go func(threadId int) {
			for j := 0; j < num; j++ {
				if j%downloadThreads == threadId {
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

	newfile, err := os.Create(outfile)
	if err != nil {
		panic(err)
	}
	defer newfile.Close()

	successNum := 0
	errorNum := 0
	// merge csv
	for i := 0; i < num; i++ {
		log.Println("parallel down load file to ", outfile, " ", i, "/", num)
		b := <-bufferCh
		if b == nil {
			errorNum += 1
			log.Println("has a fail in download one pod file ")
			continue
		}

		//only one table top is needed
		if successNum > 0 {
			_, err := b.ReadString('\n')
			if err != nil {
				panic(err)
			}
		}

		_, err = b.WriteTo(newfile)
		if err != nil {
			panic(err)
		}
		successNum++
	}

	log.Printf("download finished fail/all = %d / %d \n", errorNum, num)
}

func StopAll(cli *k8s.K8sClient) {
	pods := cli.GetPodsWithPrefix("simds")
	for _, pod := range pods {
		_, err := cli.Exec(pod, "c1", []string{"sh", "-c", "pkill simlet"}, nil, nil)
		if err != nil {
			panic(err)
		}
	}
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
	StopAll(cli)
	CollectResult(cli)
}
