package k8s

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/yaml"
)

type K8sClient struct {
	configFile      string
	clientset       *kubernetes.Clientset
	config          *restclient.Config
	namespace       string
	PodTemplate     *apiv1.Pod
	ServiceTemplate *apiv1.Service
}

// for simlet
// create readonly in container pod // for simlet
func CreateReadonlyInContainerClient() (*K8sClient, error) {
	var cli K8sClient
	// 使用InClusterConfig从默认位置获取配置
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// 创建客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.New("can not connect to the k8s in the container")
	}

	cli.clientset = clientset

	namespace, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return nil, errors.New("can not get namespace of this pod from path:  /var/run/secrets/kubernetes.io/serviceaccount/namespace")
	}
	cli.namespace = string(namespace)

	return &cli, nil
}

// for simctl
// connecting to the k8s cluster with config, this client can read and write to k8s.
func ConnectToK8s(cfgFile string, podTemplate string, serviceTemplate string) (*K8sClient, error) {
	var cli K8sClient

	if cfgFile != "" {
		cli.configFile = cfgFile
	} else if home := homedir.HomeDir(); home != "" {
		cli.configFile = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", cli.configFile)
	if err != nil {
		return nil, err
	}

	// create the clientset
	cliset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	cli.clientset = cliset
	cli.config = config

	// initialize the PodTemplate ,and ServiceTemplate
	cli.initTemplate(podTemplate, serviceTemplate)

	cli.namespace = cli.PodTemplate.Namespace
	return &cli, nil
}

func (cli *K8sClient) GetNamespace() string {
	return cli.namespace

}

func (cli *K8sClient) GetApiServerIP() string {
	u, err := url.Parse(cli.config.Host)
	if err != nil {
		panic(err)
	}
	return u.Hostname()
}

func (cli *K8sClient) initTemplate(podTemplate, serviceTemplate string) {
	data, err := os.ReadFile(podTemplate)
	if err != nil {
		panic(err)
	}
	tempPod := &apiv1.Pod{}
	err = yaml.Unmarshal(data, tempPod)
	if err != nil {
		panic(err)
	}
	cli.PodTemplate = tempPod

	data, err = os.ReadFile(serviceTemplate)
	if err != nil {
		panic(err)
	}
	tempService := &apiv1.Service{}
	err = yaml.Unmarshal(data, tempService)
	if err != nil {
		panic(err)
	}
	cli.ServiceTemplate = tempService

}

// get the k8s server ip
func (cli *K8sClient) GetServerHostName() string {
	u, err := url.Parse(cli.config.Host)
	if err != nil {
		panic(err)
	}
	return u.Hostname()
}

// create one pod in the k8s cluster
func (cli *K8sClient) CreatePod(name, lable, image string, command []string) {

	podClient := cli.clientset.CoreV1().Pods(cli.PodTemplate.Namespace)
	newPod := cli.PodTemplate.DeepCopy()
	newPod.ObjectMeta.Name = name
	newPod.ObjectMeta.Labels["app"] = lable
	newPod.Spec.Containers[0].Name = "c1"
	newPod.Spec.Containers[0].Image = image
	newPod.Spec.Containers[0].Command = command

	_, err := podClient.Create(context.Background(), newPod, metav1.CreateOptions{})

	if err != nil {
		panic(err)
	}
}

// create one pod in the k8s cluster,with reqeust, cpu : micro core, ram MiB
func (cli *K8sClient) CreatePodWithResource(name, lable, image string, command []string, cpu int64, ram int64) {

	podClient := cli.clientset.CoreV1().Pods(cli.PodTemplate.Namespace)
	newPod := cli.PodTemplate.DeepCopy()
	newPod.ObjectMeta.Name = name
	newPod.ObjectMeta.Labels["app"] = lable
	newPod.Spec.Containers[0].Name = "c1"
	newPod.Spec.Containers[0].Image = image
	newPod.Spec.Containers[0].Command = command

	newPod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU] = *resource.NewMilliQuantity(cpu, resource.DecimalSI)
	newPod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory] = *resource.NewQuantity(ram*(1024*1024), resource.DecimalSI)

	newPod.Spec.Containers[0].Resources.Limits[apiv1.ResourceCPU] = *resource.NewMilliQuantity(cpu, resource.DecimalSI)
	newPod.Spec.Containers[0].Resources.Limits[apiv1.ResourceMemory] = *resource.NewQuantity(ram*(1024*1024), resource.DecimalSI)

	_, err := podClient.Create(context.Background(), newPod, metav1.CreateOptions{})

	if err != nil {
		panic(err)
	}
}

// create one pod in the k8s cluster,with reqeust, cpu : micro core, ram MiB
func (cli *K8sClient) CreatePodWithResourceWithContainers(name, lable, image string, command []string, containerNum int, cpu int64, ram int64) {

	podClient := cli.clientset.CoreV1().Pods(cli.PodTemplate.Namespace)
	newPod := cli.PodTemplate.DeepCopy()
	newPod.ObjectMeta.Name = name
	newPod.ObjectMeta.Labels["app"] = lable
	newPod.Spec.Containers[0].Name = "c1"
	newPod.Spec.Containers[0].Image = image
	newPod.Spec.Containers[0].Command = command

	newPod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU] = *resource.NewMilliQuantity(cpu, resource.DecimalSI)
	newPod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory] = *resource.NewQuantity(ram*(1024*1024), resource.DecimalSI)

	newPod.Spec.Containers[0].Resources.Limits[apiv1.ResourceCPU] = *resource.NewMilliQuantity(cpu, resource.DecimalSI)
	newPod.Spec.Containers[0].Resources.Limits[apiv1.ResourceMemory] = *resource.NewQuantity(ram*(1024*1024), resource.DecimalSI)

	for i := 1; i < containerNum; i++ {
		newContainer := newPod.Spec.Containers[0].DeepCopy()
		newContainer.Name = fmt.Sprintf("c%d", i+1)
		newPod.Spec.Containers = append(newPod.Spec.Containers, *newContainer)
	}

	_, err := podClient.Create(context.Background(), newPod, metav1.CreateOptions{})

	if err != nil {
		panic(err)
	}
}

// create one pod in the k8s cluster
func (cli *K8sClient) DeletePod(name string) {
	podClient := cli.clientset.CoreV1().Pods(cli.PodTemplate.Namespace)
	var deleteTime int64 = 0
	err := podClient.Delete(context.Background(), name, metav1.DeleteOptions{GracePeriodSeconds: &deleteTime})
	if err != nil {
		panic(err)
	}
}

func (cli *K8sClient) CreateClusterIPService(name, lable string, inPort int) {
	svcClient := cli.clientset.CoreV1().Services(cli.ServiceTemplate.Namespace)

	newSvc := cli.ServiceTemplate.DeepCopy()
	newSvc.ObjectMeta.Name = name
	newSvc.Spec.Type = v1.ServiceTypeClusterIP
	newSvc.Spec.Selector["app"] = lable
	newSvc.Spec.Ports[0].Port = int32(inPort)
	newSvc.Spec.Ports[0].TargetPort = intstr.FromInt(inPort)

	_, err := svcClient.Create(context.Background(), newSvc, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
}

func (cli *K8sClient) CreateNodePortService(name, lable string, inPort, outPort int) {
	svcClient := cli.clientset.CoreV1().Services(cli.ServiceTemplate.Namespace)

	newSvc := cli.ServiceTemplate.DeepCopy()
	newSvc.ObjectMeta.Name = name
	newSvc.Spec.Type = v1.ServiceTypeNodePort
	newSvc.Spec.Selector["app"] = lable
	newSvc.Spec.Ports[0].Port = int32(inPort)
	newSvc.Spec.Ports[0].TargetPort = intstr.FromInt(inPort)
	newSvc.Spec.Ports[0].NodePort = int32(outPort)

	_, err := svcClient.Create(context.Background(), newSvc, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
}

// delete the service by name
func (cli *K8sClient) DeleteService(name string) {
	podClient := cli.clientset.CoreV1().Services(cli.ServiceTemplate.Namespace)
	err := podClient.Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}
}

// exec a command in a pod
func (cli *K8sClient) Exec(podName, containerName string, command []string, stdout io.Writer) error {
	clientset, config := cli.clientset, cli.config

	errBuf := &bytes.Buffer{}
	request := clientset.CoreV1().RESTClient().
		Post().
		Namespace(cli.GetNamespace()).
		Resource("pods").
		Name(podName).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Command: command,
			Stdin:   false,
			Stdout:  stdout != nil,
			Stderr:  true,
			TTY:     true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", request.URL())
	if err != nil {
		return err
	}
	err = exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdout: stdout,
		Stderr: errBuf,
	})

	if err != nil {
		return errors.New("error when exec cmd:\n" + err.Error() + "\nAnd stderr is:\n" + errBuf.String())
	}

	return nil
}

// download a file from one pod
func (cli *K8sClient) DownloadTextFile(podName, containerName, pathToCopy string, writer io.Writer) error {
	lines := int64(0)
	countLinesCmd := []string{"sh", "-c", "wc -l < " + pathToCopy}

	buf := &bytes.Buffer{}
	err := cli.Exec(podName, containerName, countLinesCmd, buf)
	if err != nil {
		return errors.New("get file lins number fail:" + err.Error())
	}
	scanner := bufio.NewScanner(buf)
	scanner.Split(bufio.ScanLines)
	scanner.Scan()
	linestr := scanner.Text()
	lines, err = strconv.ParseInt(linestr, 10, 32)

	if lines < 10000 {
		command := []string{"cat", pathToCopy}
		if err != nil {
			return err
		}
		buf := &bytes.Buffer{}
		err := cli.Exec(podName, containerName, command, buf)
		if err != nil {
			return err
		}

		_, err = buf.WriteTo(writer)
		if err != nil {
			return err
		}
		return nil
	}

	sep := 10
	for i := 0; i < sep; i++ {
		start := int(lines) * i / sep
		end := int(lines) * (i + 1) / sep
		buf := &bytes.Buffer{}
		command := []string{"sh", "-c", fmt.Sprintf("awk 'NR>%d && NR<=%d' %s", start, end, pathToCopy)}
		err := cli.Exec(podName, containerName, command, buf)
		if err != nil {
			return err
		}

		_, err = buf.WriteTo(writer)
		if err != nil {
			return err
		}
	}

	return nil
}

// delete all the pods which has the prefix
func (cli *K8sClient) DeletePodsWithPrefix(prefix string) {
	pods, err := cli.clientset.CoreV1().Pods(cli.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, item := range pods.Items {
		podName := item.GetName()
		if strings.HasPrefix(podName, prefix) {
			cli.DeletePod(podName)
		}
	}
}

func (cli *K8sClient) GetPodIP(podname string) (string, error) {
	pod, err := cli.clientset.CoreV1().Pods(cli.namespace).Get(context.TODO(), podname, metav1.GetOptions{})
	return pod.Status.PodIP, err
}

func (cli *K8sClient) GetPodHostIP(podname string) (string, error) {
	pod, err := cli.clientset.CoreV1().Pods(cli.namespace).Get(context.TODO(), podname, metav1.GetOptions{})
	return pod.Status.HostIP, err
}

// Get all the pods which has the prefix
func (cli *K8sClient) GetPodsWithPrefix(prefix string) []string {
	pods, err := cli.clientset.CoreV1().Pods(cli.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	var res []string
	for _, item := range pods.Items {
		podName := item.GetName()
		if strings.HasPrefix(podName, prefix) {
			res = append(res, podName)
		}
	}
	return res
}

// delete all service by name with the prefix
func (cli *K8sClient) DeleteServiceWithPrefix(prefix string) {
	svcs, err := cli.clientset.CoreV1().Services(cli.namespace).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		panic(err.Error())
	}

	for _, item := range svcs.Items {
		svcName := item.GetName()
		if strings.HasPrefix(svcName, prefix) {
			cli.DeleteService(svcName)
		}
	}

}

// wait until util pods running,used after create pods if needed
func (cli *K8sClient) WaitUtilAllRunning(waitPods []string) error {
	pods := cli.clientset.CoreV1().Pods(cli.namespace)

	// scan the waitPods 's status , according the result to scan again, wait finished or  return error
	for _, waitPod := range waitPods {
		thisPodRunning := false
		for !thisPodRunning {
			res, err := pods.Get(context.TODO(), waitPod, metav1.GetOptions{})
			if err != nil {
				return errors.New("can get pods info: " + waitPod)
			}

			switch res.Status.Phase {
			case v1.PodRunning:
				thisPodRunning = true
			case v1.PodFailed:
				return errors.New("error when waitting Pod:" + waitPod)
			default:
				thisPodRunning = false
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
	return nil
}
