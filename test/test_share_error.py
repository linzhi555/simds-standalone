import simtest

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    config["NodeNum"] = 100
    config["NetLatency"] = 0.8
    config["PerformanceRate"] = 0.3

    config["NetLatencyVar"] = 1.3
    simtest.test_compose(
        config,
        [simtest.shareCluster, simtest.shareK8sCluster],
        # [simtest.shareCluster],
        "shareError",
        "StateUpdatePeriod",
        [40, 100],
    )
