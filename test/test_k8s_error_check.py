import simtest

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    config["NodeNum"] = 30

    simtest.test_compose(
        config,
        [simtest.shareCluster, simtest.shareK8sCluster],
        "shareK8s",
        "PerformanceRate",
        [1.0],
    )
