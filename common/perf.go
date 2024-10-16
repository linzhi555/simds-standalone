package common

import (
	"log"
	"os"
	"runtime/pprof"
)

var memFile *os.File

func StartPerf() {
	f, err := os.OpenFile("./cpu.prof", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)

	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}

	memFile, err = os.Create("mem.prof")
	if err != nil {
		log.Fatal(err)
	}
}

func MemProf() {
	if err := pprof.WriteHeapProfile(memFile); err != nil {
		log.Fatal(err)
	}
}

func StopPerf() {
	pprof.StopCPUProfile()
	memFile.Close()
}
