import simtest

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "trace"
    simtest.test_compose(
        config,
        [simtest.centerCluster, simtest.shareCluster, simtest.dcssCluster,],
        "MultiCluster",
        "NodeNum",
        [100, 200, 400, 800],
    )
