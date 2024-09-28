import simtest

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    config["NodeNum"] = 100
    config["NetLatency"] = "0.7"

    simtest.calibre(
        config,
        simtest.dcssCluster,
        simtest.dcssK8sCluster,
        "dcssK8s_calibre",
        "NetLatencyVar",
        # [1, 1.1, 1.2, 1.23],
        [1.23,],
    )
