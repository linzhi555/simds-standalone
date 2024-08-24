import os
from typing import List
import yaml
import sys
import argparse

sys.path.append("py")
import draw


pyFileDir = os.path.dirname(os.path.realpath(__file__))


def load_config() -> dict:
    config = {}
    with open(os.path.join(pyFileDir, "config_template.yaml"), "r") as stream:
        try:
            config = yaml.safe_load(stream)
        except yaml.YAMLError as exc:
            print(exc)
    return config


class Cluster:
    def __init__(self, name, describ, command, specialConfig=None):
        self.name = name
        self.describ = describ
        self.command = command
        self.specialConfig = specialConfig


def test_compose(
    config,
    clusters: List[Cluster],
    testname: str,
    paramsName: str,
    params: List,
    parmsLables: List[str],
    drawOnly: bool = False,
):
    if not drawOnly:
        run_compose(config, clusters, testname, paramsName, params, parmsLables)
    draw_compose(clusters, testname, paramsName, params, parmsLables)

def run_compose(
    config,
    clusters: List[Cluster],
    testname: str,
    paramsName: str,
    params: List,
    parmsLables: List[str],
):
    for param, label in zip(params, parmsLables):
        for cluster in clusters:
            configCopy = config.copy()
            configCopy[paramsName] = param
            configOut = os.path.join(pyFileDir, "config.yaml")
            targetOut = os.path.join(
                pyFileDir, "target", testname, "{}_{}".format(cluster.name, label)
            )
            if cluster.specialConfig != None:
                for k, v in cluster.specialConfig.items():
                    configCopy[k] = v
            with open(configOut, "w") as output:
                yaml.dump(configCopy, output)
            os.system(
                '{} Config={} TargetFolder="{}"'.format(
                    cluster.command, "./test/config.yaml", targetOut
                )
            )

def draw_compose(
    clusters: List[Cluster],
    testname: str,
    paramsName: str,
    params: List,
    parmsLables: List[str],
):
    for _, label in zip(params, parmsLables):
        outfolder = os.path.join(pyFileDir, "target", "all", testname, label)
        os.system("mkdir -p {outfolder}" .format(outfolder=outfolder))

        tests = []
        for cluster in clusters:
            targetOut = os.path.join(
                pyFileDir, "target", testname, "{}_{}".format(cluster.name, label)
            )
            tests.append([targetOut, "{} {} ".format(cluster.describ, label)])
        draw.draw_task_submission_rate(tests,outfolder)
        draw.draw_muilt_lantencyCurve(tests,outfolder)
        draw.draw_muilt_avg_resource(tests,outfolder)
        draw.draw_muilt_var_resource(tests,outfolder)
        draw.draw_muilt_net_busy(tests,outfolder)
        draw.draw_task_latency_CDF(tests,outfolder)

    for cluster in clusters:
        outfolder = os.path.join(pyFileDir, "target", "all", testname, cluster.name)
        os.system("mkdir -p {outfolder}".format(outfolder=outfolder))

        tests = []
        for _, label in zip(params, parmsLables):
            targetOut = os.path.join(
                pyFileDir, "target", testname, "{}_{}".format(cluster.name, label)
            )
            tests.append([targetOut, "{} {} ".format(cluster.describ, label)])
        draw.draw_task_submission_rate(tests,outfolder)
        draw.draw_muilt_lantencyCurve(tests,outfolder)
        draw.draw_muilt_avg_resource(tests,outfolder)
        draw.draw_muilt_var_resource(tests,outfolder)
        draw.draw_muilt_net_busy(tests,outfolder)
        draw.draw_task_latency_CDF(tests,outfolder)


parser = argparse.ArgumentParser(description="run net shape test")
parser.add_argument("--draw_only", dest="drawOnly", help="only draw the result")
parser.set_defaults(drawOnly=False)
args = parser.parse_args()

if __name__ == "__main__":
    centerCluster = Cluster("center", "Centralized", "make test Cluster=Center")
    shareCluster = Cluster("share", "Shared State", "make test Cluster=ShareState")
    dcssCluster = Cluster("dcss", "dcss ", "make test Cluster=Dcss")

    dcssK8sCluster = Cluster("dcssk8s", "dcssk8s ", "make k8sTest Cluster=Dcss")
    shareK8sCluster = Cluster("sharek8s", "sharek8s", "make k8sTest Cluster=ShareState")

    config_template = load_config()
    
    # test the influence of net latency

    config = config_template.copy()
    config["TaskMode"] = "onePeak"
    config["TaskNumFactor"] = 6
    test_compose(
        config,
        #[dcssCluster],
        [dcssCluster, dcssK8sCluster],
        "NetLatency",
        "NetLatency",
        [1, 2, 4],
        [
            "1ms",
            "2ms",
            "4ms",
        ],
        drawOnly=args.drawOnly,
    )

    # test the state update period
    config = config_template.copy()
    config["TaskMode"] = "onePeak"
    config["TaskNumFactor"] = 4
    test_compose(
        config,
        [shareCluster, shareK8sCluster],
        "StateUpdatePeriod",
        "StateUpdatePeriod",
        [100, 150, 300],
        [
            "100ms",
            "150ms",
            "300ms",
        ],
        drawOnly=args.drawOnly,
    )

    # test the go procs

    config = config_template.copy()
    config["TaskMode"] = "onePeak"
    config["NodeNum"] = 10000
    config["TaskNumFactor"] = 6
    test_compose(
        config,
        [dcssCluster],
        "GoProcs",
        "GoProcs",
        [1, 2, 4],
        [
            "1thread",
            "2thread",
            "4thread",
        ],
        drawOnly=args.drawOnly,
    )


