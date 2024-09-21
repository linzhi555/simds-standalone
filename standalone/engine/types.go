package engine

import (
	"log"
	"simds-standalone/config"
	"time"
)

// ZEROTIME 模拟开始的现实时间，以此作为模拟器的零点时间
var ZEROTIME time.Time = time.Now()

// 每次更新代表的时间长度
var DeltaT time.Duration = time.Second / time.Duration(config.Val.FPS)

func init() {
	log.Print("DelatT :", int64(DeltaT), "ns")
}

type Progress uint32

const FullProgress Progress = 1000000

func (p *Progress) toFloat() float32 {
	return float32(*p) / float32(FullProgress)
}

func (p *Progress) Add(percent float32) {
	*p += Progress(percent * float32(FullProgress))
}

func (p *Progress) IsFinished() bool {
	return *p >= FullProgress
}
