import simtest

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    config["NodeNum"] = 800
    config["TaskNumFactor"] = 5

    simtest.test_compose(
        config,
        [simtest.centerCluster],
        "centralizedPerformance",
        "PerformanceRate",
        [0.65, 0.7, 0.75, 0.8, 1],
    )

    config["NodeNum"] = 700
    simtest.test_compose(
        config,
        [simtest.centerCluster],
        "centralizedPerformance_700",
        "PerformanceRate",
        [0.65, 0.7, 0.75, 0.8, 1],
    )

    config["NodeNum"] = 600
    simtest.test_compose(
        config,
        [simtest.centerCluster],
        "centralizedPerformance_600",
        "PerformanceRate",
        [0.65, 0.7, 0.75, 0.8, 1],
    )
