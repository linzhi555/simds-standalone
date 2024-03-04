import os
from typing import List
import yaml
import sys
import argparse
sys.path.append("py")
import draw


def load_config() -> dict:
    testDir =  os.path.dirname(os.path.realpath(__file__))
    config = {}
    with open("{testDir}/config_template.yaml".format(testDir=testDir), "r") as stream:
        try:
            config = yaml.safe_load(stream)
        except yaml.YAMLError as exc:
            print(exc)
    return config


class Cluster():
    def __init__(self,name,describ,command,specialConfig=None):
        self.name = name
        self.describ = describ
        self.command = command
        self.specialConfig = specialConfig

def test_compose(config,clusters:List[Cluster],testname:str,paramsName:str,params:List,parmsLables:List[str],drawOnly:bool=False):
    if not drawOnly:
        run_compose(config,clusters,testname,paramsName,params,parmsLables)
    draw_compose(clusters,testname,paramsName,params,parmsLables)




def run_compose(config,clusters:List[Cluster],testname:str,paramsName:str,params:List,parmsLables:List[str]):
    for param,label in zip(params,parmsLables):
        for cluster in clusters:
            configCopy = config.copy()
            configCopy[paramsName] = param
            configOut = "graduate_paper_test/config.yaml"
            targetOut = "graduate_paper_test/target/{}/{}_{}".format(testname,cluster.name,label)
            if cluster.specialConfig != None:
                for k,v in cluster.specialConfig.items():
                    configCopy[k]=v
            with open(configOut, "w") as output:
                yaml.dump(configCopy,output)
            os.system("{} Config={} TargetFolder=\"{}\"".format(cluster.command,configOut,targetOut))


def draw_compose(clusters:List[Cluster],testname:str,paramsName:str,params:List,parmsLables:List[str]):
    for _,label in zip(params,parmsLables):
        tests = []
        for cluster in clusters:
            targetOut = "graduate_paper_test/target/{}/{}_{}".format(testname,cluster.name,label)
            tests.append([targetOut, "{} {} ".format(cluster.describ,label)])
        draw.draw_task_submission_rate(tests)
        draw.draw_muilt_lantencyCurve(tests)
        draw.draw_muilt_avg_resource (tests)
        draw.draw_muilt_var_resource (tests)
        draw.draw_muilt_net_busy (tests)
        draw.draw_task_latency_CDF(tests)
        outfolder = "graduate_paper_test/target/all/{testname}/{label}".format(testname=testname,label=label)
        os.system("mkdir -p {outfolder}  && mv *.png {outfolder}".format(outfolder=outfolder))

    for cluster in clusters:
        tests = []
        for _,label in zip(params,parmsLables):
            targetOut = "graduate_paper_test/target/{}/{}_{}".format(testname,cluster.name,label)
            tests.append([targetOut, "{} {} ".format(cluster.describ,label)])
        draw.draw_task_submission_rate(tests)
        draw.draw_muilt_lantencyCurve(tests)
        draw.draw_muilt_avg_resource (tests)
        draw.draw_muilt_var_resource (tests)
        draw.draw_muilt_net_busy (tests)
        draw.draw_task_latency_CDF(tests)
        outfolder = "graduate_paper_test/target/all/{testname}/{cluster}".format(testname=testname,cluster=cluster.name)
        os.system("mkdir -p {outfolder}  && mv *.png {outfolder}".format(outfolder=outfolder))


parser = argparse.ArgumentParser(description='run net shape test')
parser.add_argument('--draw_only', dest='drawOnly', help='only draw the result')
parser.set_defaults(drawOnly=False)
args = parser.parse_args()

if __name__ == "__main__":

    centerCluster = Cluster("center","Centralized","make test Cluster=Center")
    sharedCluster = Cluster("share","Shared State","make test Cluster=ShareState")
    sparrowCluster = Cluster("sparrow","Sparrow","make test Cluster=Sparrow")
    dcssCluster = Cluster("sparrow","Sparrow","make test Cluster=Dcss")

    allclusters = [centerCluster,sharedCluster,sparrowCluster]
        
    config = load_config()

    config_copy = config.copy()
    test_compose(config_copy,[centerCluster],"center_nodenum","NodeNum",[1000,2000,3000,4000,5000],["1k","2k","3k","4k","5k"],drawOnly=args.drawOnly)
 
    config_copy = config.copy()
    test_compose(config_copy,[centerCluster],"scheduler_performance","SchedulerPerformance",[15000,20000,25000,30000],["15000","20000","25000","30000"],drawOnly=args.drawOnly)

    config_copy = config.copy()
    test_compose(config_copy,[centerCluster],"load","TaskNumFactor",[6, 6.6, 7.2, 7.8],["1.0","1.1","1.2","1.3"],drawOnly=args.drawOnly)

    config_copy = config.copy()
    test_compose(config_copy,allclusters,"nodenum","NodeNum",[1000,3000,5000],["1k","3k","5k"],drawOnly=args.drawOnly)
 
    config_copy = config.copy()
    test_compose(config_copy,[sharedCluster],"StateUpdatePeriod","StateUpdatePeriod",[100,400,1000],["100ms","400ms","1000ms"],drawOnly=args.drawOnly)

    config_copy = config.copy()
    test_compose(config_copy,allclusters,"net_lantency","NetLatency",[1,4,12],["1ms","4ms","12ms"],drawOnly=args.drawOnly)
 

