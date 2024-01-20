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

clusters = []
clusters.append(Cluster("center","Centralized","make centerTest"))
print(clusters)


utilizations = [85,90,95,100,105]
utilizationsStr = ["{per}%".format(per=x) for x in utilizations ]
TaskNumFactors = [(x/80*6.0) for x in utilizations]

# 测试集中式集群的表现与负载压力的关系
def test_utilization():
    for factor,util in zip(TaskNumFactors,utilizationsStr):
        for cluster in clusters:
            configCopy = config.copy()
            configCopy["NodeNum"] =  4*100
            configCopy["TaskNumFactor"] = factor
            configOut = os.path.join(testDir,"config.yaml")
            targetOut = os.path.join(testDir,"target/utilization/{}_{}".format(cluster.name,util))
            if cluster.specialConfig != None:
                for k,v in cluster.specialConfig.items():
                    configCopy[k]=v
            with open(configOut, "w") as output:
                yaml.dump(configCopy,output)
            os.system("{} Config={} TargetFolder=\"{}\"".format(cluster.command,configOut,targetOut))

def draw_uliliztion_test():
    for cluster in clusters:
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

parser = argparse.ArgumentParser(description='run net shape test')
parser.add_argument('--draw_only', dest='drawOnly', help='only draw the result')
parser.set_defaults(drawOnly=False)
args = parser.parse_args()

if __name__ == "__main__":
    if not args.drawOnly:
        test_utilization()
    draw_uliliztion_test()

