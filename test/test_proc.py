import simtest

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    config["NodeNum"] = 10000
    simtest.test_compose(
        config,
        [simtest.dcssCluster],
        "GoProcs",
        "GoProcs",
        [1, 2, 3, 4, 5],
        parmsLables=[
            "1thread",
            "2thread",
            "3thread",
            "4thread",
            "5thread",
        ],
    )
