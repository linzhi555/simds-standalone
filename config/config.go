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

// Config 全局的配置,在main开始前初始化
var Val struct {
	//output
	TaskEventsLogName string
	NetEventsLogName  string

	FPS        int32
	GoProcs    int32
	ConfigPath string
	Debug      bool
	OutputDir  string
	Cluster    string
	NodeNum    int32
	NetLatency int32

	NodeCpu    int32
	NodeMemory int32

	SimulateDuration int32 // how long the simulate is,Unit :ms
	TaskMode         string
	TraceFile        string
	TaskNumFactor    float32
	TaskCpu          int32
	TaskMemory       int32
	TaskLifeTime     int32

	//only in center and shareState
	SchedulerPerformance float32

	//dcss mode
	DcssNeiborRandomP float32
	DcssNeibor        int32
	DcssDividePolicy  string

	// sparrow
	//SparrowSchedulerNumFactor float32

	//ShareState the nodeinfo update information
	StateUpdatePeriod int32

	// in deploy mode
	CleanMode bool
	//K8sServicePort  int
	K8SConfig       string
	K8STemplatePath string
	NodeName        string
	PushImageRepo   string
	PullImageRepo   string
}

func init() {

	// default config
	viper.SetDefault("TaskEventsLogName", "tasks_event.log")
	viper.SetDefault("NetEventsLogName", "network_event.log")
	viper.SetDefault("FPS", 10000)
	viper.SetDefault("GoProcs", runtime.NumCPU())
	//viper.SetDefault("K8sServicePort", 31000)

	// import the command line argumet
	configFile := pflag.StringP("configFile", "c", "./config.yaml", "the config file path")
	pflag.String("Cluster", "", "which type cluster to run,for example Dcss,Center,ShareState...")
	pflag.Bool("Debug", false, "run the cluster in debug mode")
	pflag.String("OutputDir", ".", "where to output the result files")
	pflag.String("NodeName", "", "the node name")
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
