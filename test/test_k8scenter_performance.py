import simtest
import args

if __name__ == "__main__":
    config = simtest.load_config()

    config["TaskMode"] = "onePeak"
    config["NodeNum"] = 100

    simtest.test_compose(
        config,
        [simtest.centerCluster, simtest.centerK8sCluster],
        "centerK8s",
        "PerformanceRate",
        [0.8, 0.85, 0.9, 0.95, 1],
        drawOnly=args.drawOnly,
    )
