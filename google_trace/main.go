package main

import (
	"simds-standalone/cluster/lib"
)

const (
	rawOutfile                  = "raw.log"
	basicLoadRate       float64 = 1
	tasknumCompressRate int     = 4000

	lifeRate         float64 = 0.0005
	resourceRate     float64 = 10000
	timebias         int32   = 10 // use task after $timebias Second
	maxResourceLimit int32   = 500
)

func main() {
	lib.DealRawFile(basicLoadRate, lifeRate, resourceRate, timebias, maxResourceLimit, rawOutfile, "./tasks_stream.log")
}
