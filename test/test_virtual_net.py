import simtest
import args

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    simtest.test_compose(
        config,
        [simtest.centerCluster, simtest.dcssCluster, simtest.shareCluster],
        "NetLatency",
        "NetLatency",
        [1, 4, 16],
        [
            "1ms",
            "4ms",
            "16ms",
        ],
        drawOnly=args.drawOnly,
    )
