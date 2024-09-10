import simtest
import args

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

        drawOnly=args.drawOnly,
    )
