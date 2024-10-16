import simtest

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    config["NodeNum"] = 100

    config["NetLatency"] = "4"
    simtest.calibre(
        config,
        simtest.dcssCluster,
        simtest.dcssK8sCluster,
        "dcssK8s_calibre_before",
        "NetLatencyVar",
        [0.1,],
        parmsLables=["before_calibre"],
    )

    config["NetLatency"] = "0.7"
    simtest.calibre(
        config,
        simtest.dcssCluster,
        simtest.dcssK8sCluster,
        "dcssK8s_calibre",
        "NetLatencyVar",
        # [1, 1.1, 1.2, 1.23],
        [1.23,],
        parmsLables=["after_calibre"],
    )
