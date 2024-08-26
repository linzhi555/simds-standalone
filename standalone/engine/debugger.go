package engine

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/chzyer/readline"
	lua "github.com/yuin/gopher-lua"
	"simds-standalone/config"
)

// Console模式下进行运行集群
func (engine *Engine) RunInConsole() {

	luaState := simdsLua(engine)
	defer luaState.Close()

	l, err := readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf("\033[32m>>> \033[0m"),
		HistoryFile:     "/tmp/readline.tmp",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold: true,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()
	//l.CaptureExitSignal()

	log.SetOutput(l.Stderr())
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if err := luaState.DoString(line); err != nil {
			log.Println(err)
		}
	}
}

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
		fmt.Printf("Simulator Time: %f s, UpdateFrames: %d, FPS: %d \n", float32(simulator.UpdateCount)/float32(config.Val.FPS), simulator.UpdateCount, config.Val.FPS)
		return 1
	}

	l.SetGlobal("step", l.NewFunction(step))
	l.SetGlobal("to", l.NewFunction(to))
	l.SetGlobal("time", l.NewFunction(time))
	return l
}
