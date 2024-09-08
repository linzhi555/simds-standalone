import simtest
import args

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    simtest.test_compose(
        config,
        [simtest.centerCluster, simtest.shareCluster, simtest.dcssCluster,],
        "MultiCluster",
        "NodeNum",
        [100, 200, 400, 600, 800],
        drawOnly=args.drawOnly,
    )
