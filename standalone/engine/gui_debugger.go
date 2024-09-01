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

func handler(w http.ResponseWriter, r *http.Request) {
	cmd := r.URL.Query().Get("cmd")
	log.Println("start exectute", cmd, "command by lua interrpter")

	response := map[string]string{"replyForCmd": cmd, "error": "null"}
	if err := interpreter.DoString(cmd); err != nil {
		log.Println(err)
		response["error"] = err.Error()
	} else {
		response["updateCount"] = fmt.Sprint(engine.UpdateCount)
		response["upTime"] = fmt.Sprint(engine.UpTime().Milliseconds()) + "ms"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (e *Engine) GuiDebugging() {
	interpreter = simdsLua(e)
	engine = e

	http.Handle("/", http.FileServer(http.Dir("./standalone/static"))) // 静态文件服务
	http.HandleFunc("/run", handler)
	go openBrowser("http://localhost:8079") // 在后台打开浏览器
	http.ListenAndServe(":8079", nil)       // 启动服务器
}
