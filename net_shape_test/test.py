import os
import yaml
import sys
sys.path.append("py")
import draw

config = {}
with open("net_shape_test/config_template.yaml", "r") as stream:
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
clusters.append(Cluster("share","Shared State","make shareTest"))
clusters.append(Cluster("sparrow","Sparrow","make test Cluster='Sparrow'"))

# two beehive
#clusters.append(Cluster("dcss_regular","LN BeeHive ","make dcssTest", specialConfig={'DcssNeiborRandomP':0.0}))
#clusters.append(Cluster("dcss_random","WSN Beehive","make dcssTest",specialConfig={'DcssNeiborRandomP':0.5}))

# exclude LN Beehive 
clusters.append(Cluster("dcss_random","Beehive","make dcssTest",specialConfig={'DcssNeiborRandomP':0.5}))

print(clusters)

nodeNumTests = [1,2,4,7,10]
def test_node_num():
    for nodeNum in nodeNumTests:
        for cluster in clusters:
            configCopy = config.copy()
            configCopy["NodeNum"] = nodeNum * 1000
            configOut = "net_shape_test/config.yaml"
            targetOut = "net_shape_test/target/node_num/{}_{}k".format(cluster.name,nodeNum)
            if cluster.specialConfig != None:
                for k,v in cluster.specialConfig.items():
                    configCopy[k]=v
            with open(configOut, "w") as output:
                yaml.dump(configCopy,output)
            os.system("{} Config={} TargetFolder=\"{}\"".format(cluster.command,configOut,targetOut))

def draw_node_num_test():
    for nodeNum in nodeNumTests:
        tests = []
        for cluster in clusters:
            folder = "net_shape_test/target/node_num/{}_{}k".format(cluster.name,nodeNum)
            tests.append([folder, "{c} {num}k ".format(c=cluster.describ,num=nodeNum)])
        draw.draw_task_submission_rate(tests)
        draw.draw_muilt_lantencyCurve(tests)
        draw.draw_muilt_avg_resource (tests)
        draw.draw_muilt_var_resource (tests)
        draw.draw_muilt_net_busy (tests)
        draw.draw_task_latency_CDF(tests)
        os.system("mkdir -p net_shape_test/target/all/node_num/nodes_{num}k && mv *.png net_shape_test/target/all/node_num/nodes_{num}k".format(num=nodeNum))

    for cluster in clusters:
        tests = []
        for nodeNum in nodeNumTests:
            folder = "net_shape_test/target/node_num/{}_{}k".format(cluster.name,nodeNum)
            tests.append([folder, "{c} {num}k ".format(c=cluster.describ,num=nodeNum)])
        draw.draw_task_submission_rate(tests)
        draw.draw_muilt_lantencyCurve(tests)
        draw.draw_muilt_avg_resource (tests)
        draw.draw_muilt_var_resource (tests)
        draw.draw_muilt_net_busy (tests)
        draw.draw_task_latency_CDF(tests)
        os.system("mkdir -p net_shape_test/target/all/node_num/{c} && mv *.png net_shape_test/target/all/node_num/{c}".format(c = cluster.name))



possblities = [0.0,0.05,0.1,0.3,0.6,1.0]
def test_neibor_random_P():
    for p in possblities:
        cluster = clusters[3] # dcss_random
        configCopy = config.copy()
        configCopy["NodeNum"] =  4 * 1000
        configCopy["DcssNeiborRandomP"] = p
        configCopy["TaskNumFactor"] = 7.0
        configOut = "net_shape_test/config.yaml"
        targetOut = "net_shape_test/target/neibor_random_p/randomP_{}".format(p)
        with open(configOut, "w") as output:
            yaml.dump(configCopy,output)
        os.system("{} Config={} TargetFolder=\"{}\"".format(cluster.command,configOut,targetOut))

def draw_neibor_random_P_test():
    tests = []
    for p in possblities:
        cluster = clusters[3]
        folder = "net_shape_test/target/neibor_random_p/randomP_{}".format(p)
        label = ""
        if p < 0.001 :
            label = "Lattice Network"
        else :
            label = "WS Network p={}".format(p)
        tests.append([folder, label])
    draw.draw_task_submission_rate(tests)
    draw.draw_muilt_lantencyCurve(tests)
    draw.draw_muilt_avg_resource (tests)
    draw.draw_muilt_var_resource (tests)
    draw.draw_muilt_net_busy (tests)
    draw.draw_task_latency_CDF(tests)
    os.system("mkdir -p net_shape_test/target/all/neibor_random_p && mv *.png net_shape_test/target/all/neibor_random_p")

neibornum = [4,8,12,16,20]
def test_neibor_num():
    for num in neibornum:
        cluster = clusters[3] # dcss_random
        configCopy = config.copy()
        configCopy["NodeNum"] =  4 * 1000
        configCopy["DcssNeiborRandomP"] = 0.5
        configCopy["DcssNeibor"] = num
        configCopy["TaskNumFactor"] = 7.0
        configOut = "net_shape_test/config.yaml"
        targetOut = "net_shape_test/target/neibor_num/neibor_{}".format(num)
        with open(configOut, "w") as output:
            yaml.dump(configCopy,output)
        os.system("{} Config={} TargetFolder=\"{}\"".format(cluster.command,configOut,targetOut))

def draw_neibor_num():
    tests = []
    for num in neibornum:
        cluster = clusters[3]
        folder = "net_shape_test/target/neibor_num/neibor_{}".format(num)
        label = "Neighbor={}".format(num)
        tests.append([folder, label])
    draw.draw_task_submission_rate(tests)
    draw.draw_muilt_lantencyCurve(tests)
    draw.draw_muilt_avg_resource (tests)
    draw.draw_muilt_var_resource (tests)
    draw.draw_muilt_net_busy (tests)
    draw.draw_task_latency_CDF(tests)
    os.system("mkdir -p net_shape_test/target/all/neibor_num && mv *.png net_shape_test/target/all/neibor_num")



#TaskNumFactors = [6.0,6.2,6.4,6.6,6.8]
#utilizations = ["80%","83%","85%","88%","90%"]

#def test_utilization():
#    for factor,util in zip(TaskNumFactors,utilizations):
#        for cluster in clusters:
#            configCopy = config.copy()
#            configCopy["NodeNum"] =  4 * 1000
#            configCopy["TaskNumFactor"] = factor
#            configOut = "net_shape_test/config.yaml"
#            targetOut = "net_shape_test/target/utilization/{}_{}".format(cluster.name,util)
#            if cluster.specialConfig != None:
#                for k,v in cluster.specialConfig.items():
#                    configCopy[k]=v
#            with open(configOut, "w") as output:
#                yaml.dump(configCopy,output)
#            os.system("{} Config={} TargetFolder=\"{}\"".format(cluster.command,configOut,targetOut))
#
#def draw_uliliztion_test():
#    for cluster in clusters:
#        tests = []
#        for factor,ulti in zip(TaskNumFactors,utilizations):
#            folder = "net_shape_test/target/utilization/{}_{}".format(cluster.name,ulti)
#            tests.append([folder, "{c}  ".format(c=cluster.describ)])
#        draw.draw_task_submission_rate(tests)
#        draw.draw_muilt_lantencyCurve(tests)
#        draw.draw_muilt_avg_resource (tests)
#        draw.draw_muilt_var_resource (tests)
#        draw.draw_muilt_net_busy (tests)
#        draw.draw_task_latency_CDF(tests)
#        os.system("mkdir -p net_shape_test/target/all/utilization/{} && mv *.png net_shape_test/target/all/utilization/{}".format(cluster.name,cluster.name))

#def test_with_parameters(paramsName,paramsList):
#    for param in paramsList:
#        for cluster in clusters:
#            configCopy = config.copy()
#            configCopy[paramsName] = param
#            configOut = "net_shape_test/config.yaml"
#            targetOut = "net_shape_test/target/{}/{}_{}".format(paramsName,cluster.name,param)
#            if cluster.specialConfig != None:
#                for k,v in cluster.specialConfig.items():
#                    configCopy[k]=v
#            with open(configOut, "w") as output:
#                yaml.dump(configCopy,output)
#            os.system("{} Config={} TargetFolder=\"{}\"".format(cluster.command,configOut,targetOut))
#
#def draw_test_result(paramsName,paramsList):
#    for param in paramsList:
#        tests = []
#        for cluster in clusters:
#            folder = "net_shape_test/target/{}/{}_{}".format(paramsName,cluster.name,param)
#            tests.append([folder, "{} {}".format(cluster.name,param)])
#        draw.draw_task_submission_rate(tests)
#        draw.draw_muilt_lantencyCurve(tests)
#        draw.draw_muilt_avg_resource (tests)
#        draw.draw_muilt_var_resource (tests)
#        draw.draw_muilt_net_busy (tests)
#        draw.draw_task_latency_CDF(tests)
#        os.system("mkdir -p net_shape_test/target/{}/all/ && mv *.png net_shape_test/target/{}/all/".format(paramsName,paramsName))

    #test_with_parameters("TaskNumFactor",utilizationTests)
    #draw_test_result("TaskNumFactor",utilizationTests)

import argparse
parser = argparse.ArgumentParser(description='run net shape test')
parser.add_argument('--draw_only', dest='drawOnly', help='only draw the result')
parser.set_defaults(drawOnly=False)
args = parser.parse_args()

if __name__ == "__main__":

#    if not args.drawOnly:
#        test_neibor_num()
#    draw_neibor_num()


    if not args.drawOnly:
        test_neibor_random_P()
    draw_neibor_random_P_test()


    if not args.drawOnly:
        test_node_num()
    draw_node_num_test()


    #if not args.drawOnly:
    #    test_utilization()
    #draw_uliliztion_test()
