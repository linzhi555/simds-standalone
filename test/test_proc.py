import simtest
import args

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    config["NodeNum"] = 10
    simtest.test_compose(
        config,
        [simtest.dcssCluster],
        "GoProcs",
        "GoProcs",
        [1, 2, 4],
        [
            "1thread",
            "2thread",
            "4thread",
        ],
        drawOnly=args.drawOnly,
    )
