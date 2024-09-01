package engine

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
	"simds-standalone/config"
)

func simdsLua(simulator *Engine) *lua.LState {
	l := lua.NewState()
	step := func(L *lua.LState) int {
		lv := L.ToInt(1) /* get argument */
		simulator.UpdateNtimes(uint64(lv))
		return 1 /* number of results */
	}

	to := func(L *lua.LState) int {
		lv := L.ToInt(1) /* get argument */
		simulator.UpdateNtimes(uint64(lv) - simulator.UpdateCount)
		return 1 /* number of results */
	}

	time := func(L *lua.LState) int {
		fmt.Printf("Simulator Time: %dms , UpdateFrames: %d, FPS: %d \n", engine.UpTime().Milliseconds(), simulator.UpdateCount, config.Val.FPS)
		return 1
	}

	l.SetGlobal("step", l.NewFunction(step))
	l.SetGlobal("to", l.NewFunction(to))
	l.SetGlobal("time", l.NewFunction(time))
	return l
}
