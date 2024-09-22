import simtest

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    config["NodeNum"] = 100

    simtest.test_compose(
        config,
        [simtest.centerCluster],
        "algorithmCompare",
        "ScheduleFunc",
        ["firstFit", "lowestCPU"],
    )
