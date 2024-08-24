package k8s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
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
	PodTemplate     *apiv1.Pod
	ServiceTemplate *apiv1.Service
}

// connecting to the k8s cluster
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
	return &cli, nil
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
	podClient.Delete(context.Background(), name, metav1.DeleteOptions{GracePeriodSeconds: &deleteTime})
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
	podClient.Delete(context.Background(), name, metav1.DeleteOptions{})
}

// exec a command in a pod
func (cli *K8sClient) Exec(podName, containerName string, command []string, stdin io.Reader, stdout io.Writer) ([]byte, error) {

	clientset, config := cli.clientset, cli.config

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(cli.PodTemplate.Namespace).
		SubResource("exec")
	scheme := runtime.NewScheme()
	if err := apiv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("error adding to scheme: %v", err)
	}

	parameterCodec := runtime.NewParameterCodec(scheme)
	req.VersionedParams(&apiv1.PodExecOptions{
		Command:   command,
		Container: containerName,
		Stdin:     stdin != nil,
		Stdout:    stdout != nil,
		Stderr:    true,
		TTY:       false,
	}, parameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("error while creating Executor: %v", err)
	}

	var stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		return stderr.Bytes(), fmt.Errorf("error in Stream: %v", err)
	}

	return nil, nil

}

// download a file from one pod
func (cli *K8sClient) Download(podName, containerName, pathToCopy string, writer io.Writer) error {

	command := []string{"cat", pathToCopy}

	attempts := 3
	attempt := 0
	for attempt < attempts {
		attempt++

		stderr, err := cli.Exec(podName, containerName, command, nil, writer)
		if attempt == attempts {
			if len(stderr) != 0 {
				return fmt.Errorf("STDERR: " + (string)(stderr))
			}
			if err != nil {
				return err
			}
		}
		if err == nil {
			return nil
		}
	}

	return nil
}

// delete all the pods which has the prefix
func (cli *K8sClient) DeletePodsWithPrefix(prefix string) {
	pods, err := cli.clientset.CoreV1().Pods(cli.PodTemplate.Namespace).List(context.TODO(), metav1.ListOptions{})
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

func (cli *K8sClient) GetPodIP(podname string) string {
	pod, _ := cli.clientset.CoreV1().Pods(cli.PodTemplate.Namespace).Get(context.TODO(), podname, metav1.GetOptions{})
	return pod.Status.PodIP
}

func (cli *K8sClient) GetPodHostIP(podname string) string {
	pod, _ := cli.clientset.CoreV1().Pods(cli.PodTemplate.Namespace).Get(context.TODO(), podname, metav1.GetOptions{})
	return pod.Status.HostIP
}

// Get all the pods which has the prefix
func (cli *K8sClient) GetPodsWithPrefix(prefix string) []string {
	pods, err := cli.clientset.CoreV1().Pods(cli.PodTemplate.Namespace).List(context.TODO(), metav1.ListOptions{})
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
	svcs, err := cli.clientset.CoreV1().Services(cli.PodTemplate.Namespace).List(context.TODO(), metav1.ListOptions{})

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
	pods := cli.clientset.CoreV1().Pods(cli.PodTemplate.Namespace)

	// scan the waitPods 's status , according the result to scan again, wait finished or  return error
	for _, waitPod := range waitPods {
		thisPodRunning := false
		for !thisPodRunning {
			res, err := pods.Get(context.TODO(), waitPod, metav1.GetOptions{})
			if err != nil {
				return err
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
