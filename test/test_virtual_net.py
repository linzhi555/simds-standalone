import simtest
import args

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    config["NodeNum"] = 300
    simtest.test_compose(
        config,
        [simtest.centerCluster, simtest.dcssCluster, simtest.shareCluster],
        "NetLatency",
        "NetLatency",
        [1, 2, 4, 8, 16],
        ["1ms", "2ms", "4ms", "8ms", "16ms",],
        drawOnly=args.drawOnly,
    )
