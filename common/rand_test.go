package common

import "testing"

func TestRandIntWithRange(t *testing.T) {
	var sum int32

	var avg int32 = 1000
	var times int32 = 100000
	var allres map[int32]int32 = make(map[int32]int32)

	for i := int32(0); i < times; i++ {
		res := RandIntWithRange(avg, 0.3)
		sum += res
		if _, ok := allres[res]; ok {
			allres[res] += 1
		} else {
			allres[res] -= 1
		}
	}

	for i := int32(float32(avg) * 0.4); i < int32(float32(avg)*0.7); i++ {
		if _, ok := allres[i]; ok {
			t.Error("the range is too big")
		}
	}
	for i := int32(float32(avg) * 1.31); i < int32(float32(avg)*1.7); i++ {
		if _, ok := allres[i]; ok {
			t.Error("the range is too big")
		}
	}

	for i := int32(float32(avg) * 0.70); i < int32(float32(avg)*1.30); i++ {
		if count, ok := allres[i]; !ok {
			t.Error("the rand can not cover all range")
		} else {
			t.Log(i, count)
		}
	}

	var expect int32 = avg * times

	diff := float32(sum-expect) / float32(expect)

	if diff > 0.01 || diff < -0.01 {
		t.Error("the rand not correct")
	}

}
