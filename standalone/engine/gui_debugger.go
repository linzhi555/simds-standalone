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

var debuggerEngine *Engine
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
	NetState    NetDebugInfo     `json:"netState"`
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
		res.UpdateCount = fmt.Sprint(debuggerEngine.UpdateCount)
		res.UpTime = fmt.Sprint(debuggerEngine.UpTime())
		res.NodesState = debuggerEngine.DebugNodes()
		res.NetState = debuggerEngine.DebugNet()
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Println(err)
	}
}

func (e *Engine) GuiDebugging() {
	interpreter = simdsLua(e)
	debuggerEngine = e

	http.Handle("/", http.FileServer(http.Dir("./standalone/static"))) // 静态文件服务
	http.HandleFunc("/run", handler)
	go openBrowser("http://localhost:8079")  // 在后台打开浏览器
	err := http.ListenAndServe(":8079", nil) // 启动服务器
	if err != nil {
		panic(err)
	}
}
