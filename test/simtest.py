import os
from typing import List
import yaml
import sys
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


centerCluster = Cluster("center", "Centralized",
                        "make test Cluster=Center")

shareCluster = Cluster("share", "Shared State",
                       "make test Cluster=ShareState")

dcssCluster = Cluster("dcss", "dcss ", "make test Cluster=Dcss")

centerK8sCluster = Cluster("centerk8s", "centerk8s ",
                           "make k8sTest Cluster=Center")


dcssK8sCluster = Cluster("dcssk8s", "dcssk8s ",
                         "make k8sTest Cluster=Dcss")

shareK8sCluster = Cluster("sharek8s", "sharek8s",
                          "make k8sTest Cluster=ShareState")


def test_compose(
    config,
    clusters: List[Cluster],
    testname: str,
    paramsName: str,
    params: List,
    drawOnly: bool = False,
    parmsLables: List[str] = None,
):
    if parmsLables is None:
        parmsLables = []
        for param in params:
            parmsLables.append(str(param))

    if not drawOnly:
        _run_compose(config, clusters, testname, paramsName, params,
                     parmsLables)
    _draw_compose(clusters, testname, params, parmsLables)


def _run_compose(
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
            targetOut = os.path.join(pyFileDir, "target", testname,
                                     "{}_{}".format(cluster.name, label))
            if cluster.specialConfig is not None:
                for k, v in cluster.specialConfig.items():
                    configCopy[k] = v
            with open(configOut, "w") as output:
                yaml.dump(configCopy, output)

            cmd = '{} Config={} TargetFolder="{}"'.format(
                cluster.command, "./test/config.yaml", targetOut)
            retcode = os.system(cmd)
            if retcode != 0:
                print("Error when run the cmd:")
                print(cmd)
                sys.exit(0)


def _draw_compose(
    clusters: List[Cluster],
    testname: str,
    params: List,
    parmsLables: List[str],
):

    prefix = os.path.join(pyFileDir, "target", "all", testname,)

    for _, label in zip(params, parmsLables):
        outfolder = os.path.join(prefix, label)
        os.system("mkdir -p {outfolder}".format(outfolder=outfolder))

        tests = []
        for cluster in clusters:
            targetOut = os.path.join(pyFileDir, "target", testname,
                                     "{}_{}".format(cluster.name, label))
            tests.append([targetOut, "{} {} ".format(cluster.describ, label)])
        _draw_one_folder(tests, outfolder)
    for cluster in clusters:
        outfolder = os.path.join(prefix, cluster.name)
        os.system("mkdir -p {outfolder}".format(outfolder=outfolder))

        tests = []
        for _, label in zip(params, parmsLables):
            targetOut = os.path.join(pyFileDir, "target", testname,
                                     "{}_{}".format(cluster.name, label))
            tests.append([targetOut, "{} {} ".format(cluster.describ, label)])
        _draw_one_folder(tests, outfolder)

    print("draw finished:", testname, params)
    print("result dir:", prefix)


def _draw_one_folder(tests, outfolder):
    # draw.draw_task_submission_rate(tests, outfolder)
    draw.draw_muilt_lantencyCurve(tests, outfolder)
    draw.draw_muilt_avg_resource(tests, outfolder)
    draw.draw_muilt_var_resource(tests, outfolder)
    draw.draw_muilt_net_busy(tests, outfolder)
    draw.draw_task_latency_CDF(tests, outfolder)
