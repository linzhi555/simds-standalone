package common

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
)

func ParseAddr(addr string) (host, port string) {
	temp := strings.Split(addr, ":")
	if len(temp) != 2 {
		panic("addr is wrong:" + addr)
	}
	return temp[0], temp[1]
}

func IsSameHost(addr0, addr1 string) bool {
	host1, _ := ParseAddr(addr0)
	host2, _ := ParseAddr(addr1)
	return host1 == host2
}

func AssertTypeIsNotPointer(v interface{}) {
	typestr := fmt.Sprint(reflect.TypeOf(v))
	if strings.HasPrefix(typestr, "*") {
		panic(typestr + " is ponter type")
	}
}

func shuffleStringSlice(slice []string) {
	rand.Shuffle(len(slice), func(i, j int) { slice[i], slice[j] = slice[j], slice[i] })
}
