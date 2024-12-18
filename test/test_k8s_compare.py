import simtest

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    config["NodeNum"] = 100

    simtest.test_compose(
        config,
        [simtest.centerCluster, simtest.centerK8sCluster],
        "centerK8s",
        "PerformanceRate",
        [1.0],
    )

    simtest.test_compose(
        config,
        [simtest.dcssCluster, simtest.dcssK8sCluster],
        "dcssK8s",
        "PerformanceRate",
        [1.0],
    )
