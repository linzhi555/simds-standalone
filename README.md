# 分布式集群模拟器 Simds 

simds (simlulator of distributed system)
采用单机模拟与实际部署结合的方式模拟分析分布式集群的运行情况。

## 依赖
- Go >= 1.21 
- ``` go mod download ```  安装依赖
- GNU/make
- python3

- docker  (only for k8s测试,用于打包simlet镜像)
- k8s cluster config file (only for k8s实际测试，用于运行simlet集群)

##   运行离散时间驱动的单机模拟器 

- ``` make test Cluster=XXX ``` XXX = Center | ShareState | Dcss

- 根据需要修改config.yaml 内容
- 分析结果在 ./target/{实验完成时刻}/ 

- Debugger模式

- ``` make debug Cluster=XXX ``` XXX = Center | ShareState | Dcss
浏览器会自动打开调试控制台，如果没有请手动打开 ```127.0.0.1:8079```


## 运行K8S实际模拟器
config.yaml 修改 k8s config path
config.yaml 修改 镜像上传以及拉取地址
可选：修改pod_template 的namespace 为使用的namespace (默认是default)

确保部署的容器能够访问k8s api, [创建权限说明](https://kubernetes.io/zh-cn/docs/tasks/run-application/access-api-from-pod/)
命名空间为default的pod 的示例：

```
kubectl create clusterrolebinding default-view --clusterrole=view --serviceaccount=default:default
```

最终运行
``` make k8sTest Cluster=XXX ``` XXX = Center | ShareState | Dcss 

## 运行组合型测试

```
${your_python_bin} ./test/test.py 
```
如果缺少包，请用你的包管理工具(pip or conda)安装

结果在./test/target


## 模拟器原理介绍
1. 抽象层
本模拟器基于actor模型(./cluster中定义), 每个actor在消息处理的时候改变自身状态并在结束消息处理后向外发出消息（异步非阻塞）
定义集群所有的actor的消息处理方式即完成集群的定义。

2. 模拟器1: 单机运行
离散时间的系统动态模拟器 stand-alone

3. 模拟器2: 集群运行
基于k8s 将actor转换成容器部署形成容器集群 simctl 负责模拟流程，管理simlet容器， simlet 负责实例化单个actor模型。




