package common

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strings"
)

// func ParseAddr(addr string) (host, port string) {
// 	temp := strings.Split(addr, ":")
// 	if len(temp) != 2 {
// 		panic("addr is wrong:" + addr)
// 	}
// 	return temp[0], temp[1]
// }

// func IsSameHost(addr0, addr1 string) bool {
// 	host1, _ := ParseAddr(addr0)
// 	host2, _ := ParseAddr(addr1)
// 	return host1 == host2
// }

func AssertTypeIsNotPointer(v interface{}) {
	typestr := fmt.Sprint(reflect.TypeOf(v))
	if strings.HasPrefix(typestr, "*") {
		panic(typestr + " is ponter type")
	}
}

func ShuffleStringSlice(slice []string) {
	rand.Shuffle(len(slice), func(i, j int) { slice[i], slice[j] = slice[j], slice[i] })
}

func RandIntWithRange(avg int32, randRange float32) int32 {
	diff := int32(float32(avg) * randRange)
	min := avg - diff
	return min + int32(rand.Intn(int(2*diff+1)))
}

func RandInt64WithRange(avg int64, randRange float64) int64 {
	diff := int64(float64(avg) * randRange)
	min := avg - diff
	return min + int64(rand.Int63n(int64(2*diff+1)))
}

// return a expontial rand float , result is between[0, +inf)
// expontial distributis's PDF is F(x) = 1-e^(-1*lamda*x),
// average is 1/lamda, variance is 1 / (lamda)^2
// wikipedia: https://en.wikipedia.org/wiki/Exponential_distribution
func ExponentialRand(avg float64) float64 {
	x := rand.Float64()

	oneDivideLamda := float64(avg) // oneDivideLamda means "1 / lamda"

	return -1 * oneDivideLamda * math.Log(1-x)
}
