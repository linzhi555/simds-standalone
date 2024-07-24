package tasks

import (
	"math"

	lua "github.com/yuin/gopher-lua"
)

func NewLuaVm() *lua.LState {
	l := lua.NewState()
	l.SetGlobal("double", l.NewFunction(Double))
	l.SetGlobal("isPrime", l.NewFunction(IsPrime))

	return l
}

func Double(L *lua.LState) int {
	lv := L.ToInt(1) /* get argument */
	L.Push(lua.LNumber(lv * 2))
	return 1 /* number of results */
}
func IsPrime(L *lua.LState) int {
	n := L.ToInt(1) /* get argument */
	L.Push(lua.LBool(isPrime(n)))
	return 1
}

func isPrime(n int) bool {
	// 边界情况
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}

	// 判断奇数
	sqrt := int(math.Sqrt(float64(n)))
	for i := 3; i <= sqrt; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}
