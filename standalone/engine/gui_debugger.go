package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	lua "github.com/yuin/gopher-lua"
)

var engine *Engine
var interpreter *lua.LState

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		err = nil
	}
	if err != nil {
		panic(err)
	}
}

type Response struct {
	ReplyForCmd string           `json:"replyForCmd"`
	Error       string           `json:"error"`
	UpdateCount string           `json:"updateCount"`
	UpTime      string           `json:"upTime"`
	NodesState  []ActorDebugInfo `json:"nodesState"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	cmd := r.URL.Query().Get("cmd")
	log.Println("start exectute", cmd, "command by lua interrpter")

	var res Response
	if err := interpreter.DoString(cmd); err != nil {
		log.Println(err)
		res.Error = err.Error()
	} else {
		res.Error = "null"
		res.UpdateCount = fmt.Sprint(engine.UpdateCount)
		res.UpTime = fmt.Sprint(engine.UpTime().Milliseconds()) + "ms"
		res.NodesState = engine.DebugNodes()
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Println(err)
	}
}

func (e *Engine) GuiDebugging() {
	interpreter = simdsLua(e)
	engine = e

	http.Handle("/", http.FileServer(http.Dir("./standalone/static"))) // 静态文件服务
	http.HandleFunc("/run", handler)
	go openBrowser("http://localhost:8079") // 在后台打开浏览器
	http.ListenAndServe(":8079", nil)       // 启动服务器
}
