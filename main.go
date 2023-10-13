package main

import (
	"flag"
	"simds-standalone/common"
)

var Dcss = flag.Bool("dcss", false, "run dcss")
var ShareState = flag.Bool("share", false, "run share state cluster")

func init() {
	flag.Parse()
}

func main() {
	common.StartPerf()
	defer common.StopPerf()

	var cluster Cluster
	if *Dcss {
		cluster = BuildDCSSCluster()
	} else if *ShareState {
		cluster = BuildShareStateCluster()

	}else{
		cluster = BuildCenterCluster()
	}
	EcsRunCluster(cluster)

}
