import simtest
import args

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    simtest.test_compose(
        config,
        [simtest.centerCluster, simtest.shareCluster, simtest.dcssCluster,],
        "NodeNum",
        "NodeNum",
        [100, 200, 400],
        [
            "100",
            "200",
            "400",
        ],
        drawOnly=args.drawOnly,
    )
