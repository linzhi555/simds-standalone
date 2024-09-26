import simtest

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    config["NodeNum"] = 20
    config["NetLatency"] = 0.7
    config["PerformanceRate"] = 0.3
    simtest.test_compose(
        config,
        # [simtest.shareCluster, simtest.shareK8sCluster],
        [simtest.shareCluster, simtest.shareK8sCluster],
        "shareError",
        "StateUpdatePeriod",
        [10, 20, 30, 40],
    )
