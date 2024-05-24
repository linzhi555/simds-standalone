import os
from typing import List
import yaml
import sys
import argparse
sys.path.append("py")
import draw


pyFileDir= os.path.dirname(os.path.realpath(__file__))


def load_config() -> dict:
    config = {}
    with open( os.path.join(pyFileDir,"config_template.yaml"), "r") as stream:
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
            configOut = os.path.join(pyFileDir, "config.yaml")
            targetOut = os.path.join(pyFileDir,"target",testname,"{}_{}".format(cluster.name,label))
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
            targetOut = os.path.join(pyFileDir,"target",testname,"{}_{}".format(cluster.name,label))
            tests.append([targetOut, "{} {} ".format(cluster.describ,label)])
        draw.draw_task_submission_rate(tests)
        draw.draw_muilt_lantencyCurve(tests)
        draw.draw_muilt_avg_resource (tests)
        draw.draw_muilt_var_resource (tests)
        draw.draw_muilt_net_busy (tests)
        draw.draw_task_latency_CDF(tests)
        outfolder = os.path.join(pyFileDir,"target","all",testname,label)
        os.system("mkdir -p {outfolder}  && mv *.png {outfolder}".format(outfolder=outfolder))

    for cluster in clusters:
        tests = []
        for _,label in zip(params,parmsLables):
            targetOut = os.path.join(pyFileDir,"target",testname,"{}_{}".format(cluster.name,label))
            tests.append([targetOut, "{} {} ".format(cluster.describ,label)])
        draw.draw_task_submission_rate(tests)
        draw.draw_muilt_lantencyCurve(tests)
        draw.draw_muilt_avg_resource (tests)
        draw.draw_muilt_var_resource (tests)
        draw.draw_muilt_net_busy (tests)
        draw.draw_task_latency_CDF(tests)
        outfolder = os.path.join(pyFileDir,"target","all", testname ,cluster.name)
        os.system("mkdir -p {outfolder}  && mv *.png {outfolder}".format(outfolder=outfolder))


parser = argparse.ArgumentParser(description='run net shape test')
parser.add_argument('--draw_only', dest='drawOnly', help='only draw the result')
parser.set_defaults(drawOnly=False)
args = parser.parse_args()

if __name__ == "__main__":

    centerCluster = Cluster("center","Centralized","make test Cluster=Center")
    sharedCluster = Cluster("share","Shared State","make test Cluster=ShareState")
    dcssCluster = Cluster("dcss","dcss ","make test Cluster=Dcss")
    
    dcssK8sCluster = Cluster("dcssk8s","dcssk8s ","make k8sTest Cluster=Dcss")

        
    config = load_config()

    config_copy = config.copy()
    test_compose(config_copy,[dcssCluster,dcssK8sCluster],"nodenum","NodeNum",[20],["20"],drawOnly=args.drawOnly)

    # config_copy = config.copy()
    # config_copy["NodeNum"] = 1000
    # test_compose(config_copy,[centerCluster],"task_mode","TaskMode",["noWave","onePeak","trace"],["constantly","pulse","trace"],drawOnly=args.drawOnly)
 
    # config_copy = config.copy()
    # test_compose(config_copy,[centerCluster],"scheduler_performance","SchedulerPerformance",[15000,20000,25000,30000],["15000","20000","25000","30000"],drawOnly=args.drawOnly)

    # config_copy = config.copy()
    # test_compose(config_copy,[centerCluster,sharedCluster,sparrowCluster],"nodenum","NodeNum",[1000,3000,5000],["1k","3k","5k"],drawOnly=args.drawOnly)
 
    # config_copy = config.copy()
    # test_compose(config_copy,[sharedCluster],"StateUpdatePeriod","StateUpdatePeriod",[100,400,1000],["100ms","400ms","1000ms"],drawOnly=args.drawOnly)

    # config_copy = config.copy()
    # test_compose(config_copy,[centerCluster,sharedCluster,sparrowCluster],"net_lantency","NetLatency",[1,4,12],["1ms","4ms","12ms"],drawOnly=args.drawOnly)
 

