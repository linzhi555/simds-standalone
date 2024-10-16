import os
from typing import List
import yaml
import sys
import draw

pyFileDir = os.path.dirname(os.path.realpath(__file__))


def load_config(filepath: str = "config_template.yaml") -> dict:
    config = {}
    with open(os.path.join(pyFileDir, filepath), "r") as stream:
        try:
            config = yaml.safe_load(stream)
        except yaml.YAMLError as exc:
            print(exc)
    return config


class Cluster:

    def __init__(self, name, describ, command,):
        self.name = name
        self.describ = describ
        self.command = command


centerCluster = Cluster("center", "central", "make test Cluster=Center")

shareCluster = Cluster("share", "share", "make test Cluster=ShareState")

dcssCluster = Cluster("dcss", "dcss", "make test Cluster=Dcss")


centerK8sCluster = Cluster("centerk8s", "central deploy ",
                           "make k8sTest Cluster=Center")


shareK8sCluster = Cluster("sharek8s", "share deploy",
                          "make k8sTest Cluster=ShareState")

dcssK8sCluster = Cluster("dcssk8s", "dcss deploy ",
                         "make k8sTest Cluster=Dcss")

# 比较不同参数的软件模拟的集群和固定参数的k8s集群部署的表现"""
# 为了校准得到，哪个参数最贴近实际情况 """


def calibre(
        config: map,
        simCluster: Cluster,
        k8sCluster: Cluster,
        testname: str,
        paramsName: str,
        params: List,
        parmsLables: List[str] = None,
):

    if parmsLables is None:
        parmsLables = []
        for param in params:
            parmsLables.append(str(param))

    testdatas = draw.testDataList()

    # 进行一次k8s实验，实验结果作为软件模拟器的校准目标
    targetOut = os.path.join(pyFileDir, "target", testname,
                             "{}".format(k8sCluster.name))
    _run_single_test(k8sCluster, config, targetOut)
    testdatas.add_data(targetOut, "{}".format(k8sCluster.describ))

    # 进行不同参数的软件模拟实验，看哪个参数更接近校准目标
    for param, label in zip(params, parmsLables):
        targetOut = os.path.join(pyFileDir, "target", testname,
                                 "{}_{}".format(simCluster.name, label))

        configCopy = config.copy()
        configCopy[paramsName] = param
        _run_single_test(simCluster, configCopy, targetOut)
        testdatas.add_data(targetOut, "{}_{}".format(
            simCluster.describ, label))

    outfolder = os.path.join(pyFileDir, "target", "all", testname,)
    os.system("mkdir -p {}".format(outfolder))
    draw.all(testdatas, outfolder)


def _run_single_test(cluster: Cluster, config: map, targetOut: str):

    if os.path.isdir(targetOut):
        print(targetOut, "already existed, so skip this test")
        return

    configOut = os.path.join(pyFileDir, "config.yaml")
    with open(configOut, "w") as output:
        yaml.dump(config, output)

    cmd = '{} Config={} TargetFolder="{}"'.format(
        cluster.command, "./test/config.yaml", targetOut)
    retcode = os.system(cmd)
    if retcode != 0:
        print("Error when run the cmd:")
        print(cmd)
        sys.exit(0)

    pass


def test_compose(
    config: map,
    clusters: List[Cluster],
    testname: str,
    paramsName: str,
    params: List,
    parmsLables: List[str] = None,
):
    if parmsLables is None:
        parmsLables = []
        for param in params:
            parmsLables.append(str(param))

    _run_compose(config, clusters, testname,
                 paramsName, params, parmsLables)

    _draw_compose(clusters, testname, params, parmsLables)


def _run_compose(
    config: map,
    clusters: List[Cluster],
    testname: str,
    paramsName: str,
    params: List,
    parmsLables: List[str],
):
    for param, label in zip(params, parmsLables):
        for cluster in clusters:

            targetOut = os.path.join(pyFileDir, "target", testname,
                                     "{}_{}".format(cluster.name, label))
            configCopy = config.copy()
            configCopy[paramsName] = param
            _run_single_test(cluster, configCopy, targetOut)


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

        tests = draw.testDataList()
        for cluster in clusters:
            dataFolder = os.path.join(pyFileDir, "target", testname,
                                      "{}_{}".format(cluster.name, label))
            tests.add_data(dataFolder, "{}".format(cluster.describ))
        draw.all(tests, outfolder)

    for cluster in clusters:
        outfolder = os.path.join(prefix, cluster.name)
        os.system("mkdir -p {outfolder}".format(outfolder=outfolder))

        tests = draw.testDataList()
        for _, label in zip(params, parmsLables):
            dataFolder = os.path.join(pyFileDir, "target", testname,
                                      "{}_{}".format(cluster.name, label))
            tests.add_data(dataFolder, "{}".format(label))
        draw.all(tests, outfolder)

    print("draw finished:", testname, params)
    print("result dir:", prefix)
