package main

import (
	"log"
	"os"
	"runtime/pprof"
)

func startPerf() {
	f, err := os.OpenFile("./cpu.prof", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
}

func stopPerf() {
	pprof.StopCPUProfile()
}
