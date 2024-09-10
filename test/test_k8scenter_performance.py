import simtest
import args

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"

    simtest.test_compose(
        config,
        [simtest.centerCluster, simtest.centerK8sCluster],
        "centerK8s",
        "NodeNum",
        [10, 40, 100],
        drawOnly=args.drawOnly,
    )
