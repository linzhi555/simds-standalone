import simtest
import args

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    config["NodeNum"] = 500

    simtest.test_compose(
        config,
        [simtest.centerCluster],
        "centeralizedPerformance",
        "PerformanceRate",
        [1, 0.8, 0.6, 0.5, 0.4],
        drawOnly=args.drawOnly,
    )
