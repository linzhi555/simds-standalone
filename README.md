# 分布式集群模拟器单机版

simds-standalone(simulator of distribute cluster - standalone edition)
可用于模拟分布式集群的任务调度，任务运行，任务信息通信等行为。同时支持与集中式集群
对比。

## 依赖

- Go >= 1.21 
- ``` go mod download ```  安装依赖
- GNU/make
- python3

- docker  (only for k8s测试)
- k8s cluster config file (only for k8s实际测试)

##   运行离散时间驱动的单机模拟器 

- ``` make test Cluster=XXX ``` XXX = Center | ShareState | Dcss

- 根据需要修改config.yaml 内容
- 分析结果在 ./target/{实验完成时刻}/ 

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
{$your_python_bin} ./test/test.py 
```
如果缺少包，请用你的包管理工具(pip or conda)安装

结果在./test/target


## 模拟器原理介绍

