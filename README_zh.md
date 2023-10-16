# 分布式集群模拟器单机版

simds-standalone(simulator of distribute cluster - standalone edition)
可用于模拟分布式集群的任务调度，任务运行，任务信息通信等行为。同时支持与集中式集群
对比。

# 使用方法

- 安装 Go >= 1.19 , make
- ``` go get ```  安装依赖
- 根据需要修改config.yaml 内容
- 运行某一个测试
    - 集中式集群测试  ``` make centerTest ```
    - 运行状态共享集群测试 ```make shareTest ```
    - 分布式集群测试 ```make dcssTest ```
- 模拟过程 所有组件的日志信息在 ./components.log
- 模拟 所有任务的提交，开始，介绍 任务日志保存在 ./tasks_event.log
- 分析结果在 ./target/{实验完成时刻}/ 


