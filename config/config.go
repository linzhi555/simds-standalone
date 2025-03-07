// config.go 统一管理配置
// 命令行能配置核心的参数配置文件位置，运行集群类型，输出文件位置
// config.yaml 配置剩余运行的参数
package config

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Gloabl Config 全局的配置, inited before doing everything
var Val struct {
	// baic
	SimulateDuration int32 // how long the simulate is,Unit :ms
	FPS              int32
	GoProcs          int32
	ConfigPath       string
	Debug            bool
	OutputDir        string
	Cluster          string

	//log
	TaskEventsLogName string
	NetEventsLogName  string

	// worker
	NodeNum    int32
	NodeCpu    int32
	NodeMemory int32

	// net enviroment
	NetLatency    float32 // net latency: Unit ms
	NetLatencyVar float32 // net latency: Unit ms

	// task generator
	TaskMode      string
	TraceFile     string
	TaskNumFactor float32
	TaskCpu       int32
	TaskMemory    int32
	TaskLifeTime  int32

	// sceduler
	ScheduleFunc string

	// machine speed ratio
	PerformanceRate float32

	// dcss mode
	DcssNeiborRandomP float32
	DcssNeibor        int32
	DcssDividePolicy  string

	// ShareState
	StateUpdatePeriod int32
	ShareSchdulerNum  int32
	StorageNum        int32

	// in deploy mode
	CleanMode       bool
	K8SConfig       string
	K8STemplatePath string
	NodeName        string
	PushImageRepo   string
	PullImageRepo   string
}

// use 'viper' to manage all config source, every config source has priority
// for viper, command argumet > config.yaml > default
func init() {

	// default config
	viper.SetDefault("TaskEventsLogName", "tasks_event.log")
	viper.SetDefault("NetEventsLogName", "network_event.log")
	viper.SetDefault("FPS", 10000)
	viper.SetDefault("GoProcs", runtime.NumCPU())

	// import the command line argumet
	configFile := pflag.StringP("configFile", "c", "./config.yaml", "the config file path")
	pflag.String("Cluster", "", "which type cluster to run,for example Dcss | Center | ShareState ...")
	pflag.Bool("Debug", false, "run the cluster in debug mode")
	pflag.String("OutputDir", ".", "where to output the result files")
	pflag.String("NodeName", "", "the node name of the k8s container, used to init first actor")
	pflag.Bool("CleanMode", false, "clean the containers")
	pflag.Parse()

	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		panic(err)
	}

	// import config from file
	Val.ConfigPath = *configFile
	viper.SetConfigFile(*configFile)
	if err := viper.ReadInConfig(); err != nil {
		panic("config file read fail: " + err.Error())
	}

	// get the final result
	err = viper.Unmarshal(&Val)
	if err != nil {
		panic(err)
	}
}

func LogString() string {
	fields := strings.FieldsFunc(fmt.Sprintf("%+v\n", Val), func(r rune) bool {
		return r == '{' || r == '}' || r == ' '
	})
	confInfo := strings.Join(fields, "\n")
	return fmt.Sprintln("config of this simulation is\n" + confInfo)
}

// 输出分析后的最终配置结果
func LogConfig(outputPath string) {
	confInfo := LogString()
	log.Println(confInfo)
	f, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err2 := f.WriteString(confInfo)

	if err2 != nil {
		log.Fatal(err2)
	}

}
