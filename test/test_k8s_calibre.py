import simtest

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    config["NodeNum"] = 30

    config["NetLatency"] = "0.7"
    config["NetLatencyVar"] = "1.22"
    simtest.test_compose(
        config,
        [simtest.dcssCluster, simtest.dcssK8sCluster],
        "dcssK8s_calibre",
        "PerformanceRate",
        [0.3],
    )
