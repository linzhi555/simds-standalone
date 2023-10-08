package main

import (
	"flag"
	"simds-standalone/common"
)

var Dcss = flag.Bool("dcss", false, "run dcss")

func init() {
	flag.Parse()
}

func main() {
	common.StartPerf()
	defer common.StopPerf()

	var cluster Cluster
	if *Dcss {
		cluster = BuildDCSSCluster()
	} else {
		cluster = BuildCenterCluster()
	}
	EcsRunCluster(cluster)

}
