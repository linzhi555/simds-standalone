package main

import (
	"flag"
	"simds-standalone/common"
)

var Debug = flag.Bool("debug", false, "run as debug mode")
var Dcss = flag.Bool("dcss", false, "run dcss")

func init() {
	flag.Parse()
}

func main() {
	common.StartPerf()
	defer common.StopPerf()

	cluster := BuildCenterCluster()
	EcsRunCluster(cluster)

}
