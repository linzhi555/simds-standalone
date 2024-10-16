package common_test

import (
	"simds-standalone/common"
	"testing"
	"time"
)

func TestUID(t *testing.T) {
	var res [1000]string

	start := time.Now()
	for i := 0; i < len(res); i++ {
		res[i] = common.GenerateUID()
	}

	cost := time.Since(start)
	if cost > 10*time.Millisecond {
		t.Error("cost to much time generating uid", cost)
	}

	resmap := map[string]struct{}{}

	for i := 0; i < len(res); i++ {
		if _, ok := resmap[res[i]]; ok {
			t.Errorf("uid duplicated")
		}
		resmap[res[i]] = struct{}{}
	}
}
