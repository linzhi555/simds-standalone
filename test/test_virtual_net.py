import simtest
import args

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    simtest.test_compose(
        config,
        [simtest.dcssCluster, simtest.dcssK8sCluster],
        "NetLatency",
        "NetLatency",
        [1, 2, 4],
        [
            "1ms",
            "2ms",
            "4ms",
        ],
        drawOnly=args.drawOnly,
    )
