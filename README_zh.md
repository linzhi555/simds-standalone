# 分布式集群模拟器单机版
simds-standalone(simulator of distribute cluster - standalone edition)
可用于模拟分布式集群的任务调度，任务运行，任务信息通信等行为。同时支持与集中式集群
对比。

# 使用方法
1. 安装 Go >= 1.19 , make 
2. 集中式集群测试 make centerTest
3. 分布式集群测试 make dcssTest
4. 如果需要则保存运行日志 make XXXTest > result.log
5. 分析结果在 target/实验完成时刻/ 


