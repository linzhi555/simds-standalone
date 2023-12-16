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
clusters.append(Cluster("center","centralized cluster","make centerTest"))
clusters.append(Cluster("share","shared cluster","make shareTest"))
clusters.append(Cluster("dcss_regular","Hive regular neighbor","make dcssTest", specialConfig={'DcssNeiborRandomP':0.0}))
clusters.append(Cluster("dcss_random","Hive random neighbor","make dcssTest",specialConfig={'DcssNeiborRandomP':0.5}))
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
            tests.append([folder, "{c} {num}k nodes".format(c=cluster.describ,num=nodeNum)])
        draw.draw_task_submission_rate(tests)
        draw.draw_muilt_lantencyCurve(tests)
        draw.draw_muilt_avg_resource (tests)
        draw.draw_muilt_var_resource (tests)
        draw.draw_muilt_net_busy (tests)
        draw.draw_task_latency_CDF(tests)
        os.system("mkdir -p net_shape_test/target/node_num/all/nodes_{num}k && mv *.png net_shape_test/target/node_num/all/nodes_{num}k".format(num=nodeNum))

possblities = [0.0,0.3,0.6,1.0]

def test_neibor_random_P():
    for p in possblities:
        cluster = clusters[3] # dcss_random
        configCopy = config.copy()
        configCopy["NodeNum"] =  4 * 1000
        configCopy["DcssNeiborRandomP"] = p
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
        tests.append([folder, "random neighbor p {} ".format(p)])
    draw.draw_task_submission_rate(tests)
    draw.draw_muilt_lantencyCurve(tests)
    draw.draw_muilt_avg_resource (tests)
    draw.draw_muilt_var_resource (tests)
    draw.draw_muilt_net_busy (tests)
    draw.draw_task_latency_CDF(tests)
    os.system("mkdir -p net_shape_test/target/all/neibor_random_p && mv *.png net_shape_test/target/all/neibor_random_p")

TaskNumFactors = [6.0,6.2,6.4,6.6,6.8]
utilizations = ["80%","83%","85%","88%","90%"]

def test_utilization():
    for factor,util in zip(TaskNumFactors,utilizations):
        for cluster in clusters:
            configCopy = config.copy()
            configCopy["TaskNumFactor"] = factor
            configOut = "net_shape_test/config.yaml"
            targetOut = "net_shape_test/target/utilization/{}_{}".format(cluster.name,util)
            if cluster.specialConfig != None:
                for k,v in cluster.specialConfig.items():
                    configCopy[k]=v
            with open(configOut, "w") as output:
                yaml.dump(configCopy,output)
            os.system("{} Config={} TargetFolder=\"{}\"".format(cluster.command,configOut,targetOut))

def draw_uliliztion_test():
    for cluster in clusters:
        tests = []
        for factor,ulti in zip(TaskNumFactors,utilizations):
            folder = "net_shape_test/target/utilization/{}_{}".format(cluster.name,ulti)
            tests.append([folder, "{c}  ".format(c=cluster.describ)])
        draw.draw_task_submission_rate(tests)
        draw.draw_muilt_lantencyCurve(tests)
        draw.draw_muilt_avg_resource (tests)
        draw.draw_muilt_var_resource (tests)
        draw.draw_muilt_net_busy (tests)
        draw.draw_task_latency_CDF(tests)
        os.system("mkdir -p net_shape_test/target/all/utilization/{} && mv *.png net_shape_test/target/all/utilization/{}".format(cluster.name,cluster.name))

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


if __name__ == "__main__":
    test_node_num()
    draw_node_num_test()

    test_neibor_random_P()
    draw_neibor_random_P_test()

    test_utilization()
    draw_uliliztion_test()
