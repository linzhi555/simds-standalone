package common

import (
	"log"
	"os"
	"runtime/pprof"
)

func StartPerf() {
	f, err := os.OpenFile("./cpu.prof", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
}

func StopPerf() {
	pprof.StopCPUProfile()
}
