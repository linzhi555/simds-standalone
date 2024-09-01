package engine
// Console模式下进行运行集群

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/chzyer/readline"
)

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
