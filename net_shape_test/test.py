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
clusters.append(Cluster("dcss_regular","dcss_regular","make dcssTest", specialConfig={'DcssNeiborRandomP':0.0}))
clusters.append(Cluster("dcss_random","dcss_random","make dcssTest",specialConfig={'DcssNeiborRandomP':0.5}))
print(clusters)

nodeNumTests = [1,2,4,7,10]
def test_node_num():
    for nodeNum in nodeNumTests:
        for cluster in clusters:
            configCopy = config.copy()
            configCopy["NodeNum"] = nodeNum * 1000
            configOut = "net_shape_test/config.yaml"
            targetOut = "net_shape_test/target/{}_{}k".format(cluster.name,nodeNum)
            if cluster.specialConfig != None:
                for k,v in cluster.specialConfig.items():
                    configCopy[k]=v
            with open(configOut, "w") as output:
                yaml.dump(configCopy,output)
            os.system("{} Config={} TargetFolder=\"{}\"".format(cluster.command,configOut,targetOut))

def draw_node_num():
    for nodeNum in nodeNumTests:
        tests = []
        for cluster in clusters:
            folder = "net_shape_test/target/{}_{}k".format(cluster.name,nodeNum)
            tests.append([folder, "{c} {num}k nodes".format(c=cluster.describ,num=nodeNum)])
        draw.draw_task_submission_rate(tests)
        draw.draw_muilt_lantencyCurve(tests)
        draw.draw_muilt_avg_resource (tests)
        draw.draw_muilt_var_resource (tests)
        draw.draw_muilt_net_busy (tests)
        draw.draw_task_latency_CDF(tests)
        os.system("mkdir -p net_shape_test/target/all/nodes_{num}k && mv *.png net_shape_test/target/all/nodes_{num}k".format(num=nodeNum))


test_node_num()
draw_node_num()

