// config.go 统一管理配置
// 命令行能配置核心的参数配置文件位置，运行集群类型，输出文件位置
// config.yaml 配置剩余运行的参数
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config 全局的配置,在main开始前初始化
var Config struct {
	OutputDir            string
	Cluster              string
	NodeNum              int32
	NetLatency           int32
	DcssNeibor           int32
	NodeCpu              int32
	NodeMemory           int32
	TaskNumFactor        float32
	TaskCpu              int32
	TaskMemory           int32
	TaskLifeTime         int32
	SchedulerPerformance float32
	StateUpdatePeriod    int32
}

func init() {

	configFile := pflag.StringP("configFile", "c", "./config.yaml", "the config file path")
	pflag.String("Cluster", "", "which type cluster to run,for example Dcss,Center,ShareState...")
	pflag.String("OutputDir", ".", "where to output the result files")
	pflag.Parse()

	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		panic(err)
	}

	viper.SetConfigFile(*configFile)
	if err := viper.ReadInConfig(); err != nil {
		panic("config file not find")
	}
	err = viper.Unmarshal(&Config)
	if err != nil {
		panic(err)
	}
}

// 输出分析后的最终配置结果
func LogConfig(outputPath string) {
	fields := strings.FieldsFunc(fmt.Sprintf("%+v\n", Config), func(r rune) bool {
		return r == '{' || r == '}' || r == ' '
	})
	confInfo := strings.Join(fields, "\n")
	log.Println("config of this simulation is\n" + confInfo)

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
