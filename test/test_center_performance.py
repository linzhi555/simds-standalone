import simtest
import args

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    simtest.test_compose(
        config,
        [simtest.centerCluster],
        "centeralizedPerformance",
        "NodeNum",
        [300, 350, 400, 450, 500],
        ["300", "350", "400", "450", "500",],
        drawOnly=args.drawOnly,
    )
