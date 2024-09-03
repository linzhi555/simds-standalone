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
        [1000, 2000, 4000],
        [
            "1k",
            "2k",
            "4k",
        ],
        drawOnly=args.drawOnly,
    )
