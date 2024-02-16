import os
import yaml
import sys
import argparse
sys.path.append("py")
import draw


testDir =  os.path.dirname(os.path.realpath(__file__))
config = {}

with open("{testDir}/config_template.yaml".format(testDir=testDir), "r") as stream:
    try:
        config = yaml.safe_load(stream)
    except yaml.YAMLError as exc:
        print(exc)


class Cluster():
    def __init__(self,name,describ,command,specialConfig=None):
        self.name = name
        self.describ = describ
        self.command = command
        self.specialConfig = specialConfig


# 保持集群规模不变
# 测试集群的表现与负载压力的关系


utilizations = [85,90,95,100,105]
utilizationsStr = ["{per}".format(per=x) for x in utilizations ]
TaskNumFactors = [(x/80*6.0) for x in utilizations]
def test_utilization(cluster):
    for factor,util in zip(TaskNumFactors,utilizationsStr):
        configCopy = config.copy()
        configCopy["NodeNum"] =  1000
        configCopy["TaskNumFactor"] = factor
        configOut = os.path.join(testDir,"config.yaml")
        targetOut = os.path.join(testDir,"target/utilization/{}_{}".format(cluster.name,util))
        if cluster.specialConfig != None:
            for k,v in cluster.specialConfig.items():
                configCopy[k]=v
        with open(configOut, "w") as output:
            yaml.dump(configCopy,output)
        os.system("{} Config={} TargetFolder=\"{}\"".format(cluster.command,configOut,targetOut))

def draw_uliliztion_test(cluster):
    tests = []
    for factor,util in zip(TaskNumFactors,utilizationsStr):
        folder = os.path.join(testDir,"target/utilization/{}_{}".format(cluster.name,util))
        tests.append([folder, "{c} {util}% ".format(c=cluster.describ,util=util)])
    draw.draw_task_submission_rate(tests)
    draw.draw_muilt_lantencyCurve(tests)
    draw.draw_muilt_avg_resource (tests)
    draw.draw_muilt_var_resource (tests)
    draw.draw_muilt_net_busy (tests)
    draw.draw_task_latency_CDF(tests)
    os.system("mkdir -p {testDir}/target/all/utilization/{cluster} && mv *.png {testDir}/target/all/utilization/{cluster}".format(testDir=testDir,cluster=cluster.name))

# 保持相对负载压力水平不平
# 测试集群的表现与集群规模的关系
nodenums = [1700,1800,1900,2000,2100]
def test_scales(cluster):
    for nodenum in nodenums:
        configCopy = config.copy()
        configCopy["NodeNum"] =  nodenum
        configCopy["TaskNumFactor"] = 6.0
        configOut = os.path.join(testDir,"config.yaml")
        targetOut = os.path.join(testDir,"target/nodenum/{}_{}".format(cluster.name,nodenum))
        with open(configOut, "w") as output:
            yaml.dump(configCopy,output)
        os.system("{} Config={} TargetFolder=\"{}\"".format(cluster.command,configOut,targetOut))


def draw_scales_test(cluster):
    tests = []
    for nodenum in nodenums:
        folder = os.path.join(testDir,"target/nodenum/{}_{}".format(cluster.name,nodenum))
        tests.append([folder, "{node}".format(node=nodenum)])
    draw.draw_task_submission_rate(tests)
    draw.draw_muilt_lantencyCurve(tests)
    draw.draw_muilt_avg_resource (tests)
    draw.draw_muilt_var_resource (tests)
    draw.draw_muilt_net_busy (tests)
    draw.draw_task_latency_CDF(tests)
    os.system("mkdir -p {testDir}/target/all/nodenum/{cluster} && mv *.png {testDir}/target/all/nodenum/{cluster}".format(testDir=testDir,cluster=cluster.name))
    return

# 测试网络延迟对集群的表现的影响

net_latencys=[1,5,9,13,17]
def test_net_latency(cluster):
    for latency in net_latencys:
        configCopy = config.copy()
        configCopy["NodeNum"] =  2000
        configCopy["NetLatency"] = latency
        configOut = os.path.join(testDir,"config.yaml")
        targetOut = os.path.join(testDir,"target/net_latency/{}_{}".format(cluster.name,latency))
        with open(configOut, "w") as output:
            yaml.dump(configCopy,output)
        os.system("{} Config={} TargetFolder=\"{}\"".format(cluster.command,configOut,targetOut))

def draw_net_latency_test(cluster):
    tests = []
    for latency in net_latencys:
        folder = os.path.join(testDir,"target/net_latency/{}_{}".format(cluster.name,latency))
        tests.append([folder, "{latency}".format(latency=latency)])
    draw.draw_task_submission_rate(tests)
    draw.draw_muilt_lantencyCurve(tests)
    draw.draw_muilt_avg_resource (tests)
    draw.draw_muilt_var_resource (tests)
    draw.draw_muilt_net_busy (tests)
    draw.draw_task_latency_CDF(tests)
    os.system("mkdir -p {testDir}/target/all/net_latency/{cluster} && mv *.png {testDir}/target/all/net_latency/{cluster}".format(testDir=testDir,cluster=cluster.name))
    return

parser = argparse.ArgumentParser(description='run net shape test')
parser.add_argument('--draw_only', dest='drawOnly', help='only draw the result')
parser.set_defaults(drawOnly=False)
args = parser.parse_args()

if __name__ == "__main__":

    centerCluster = Cluster("center","Centralized","make test Cluster=Center")
    if not args.drawOnly:
        test_utilization(centerCluster)
    draw_uliliztion_test(centerCluster)
 
    if not args.drawOnly:
        test_scales(centerCluster)
    draw_scales_test(centerCluster)
 
    sharedCluster = Cluster("share","Shared State","make test Cluster=ShareState")
    if not args.drawOnly:
        test_net_latency(sharedCluster)
    draw_net_latency_test(sharedCluster)

