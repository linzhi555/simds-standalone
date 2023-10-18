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
	Dcss                 bool
	ShareState           bool
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

	pflag.BoolP("Dcss", "d", false, "run dcss")
	pflag.BoolP("ShareState", "s", false, "run share state cluster")

	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		panic(err)
	}

	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	if err := viper.ReadInConfig(); err != nil {
		panic("config file not find")
	}
	err = viper.Unmarshal(&Config)
	if err != nil {
		panic(err)
	}

}

func LogConfig() {
	fields := strings.FieldsFunc(fmt.Sprintf("%+v\n", Config), func(r rune) bool {
		return r == '{' || r == '}' || r == ' '
	})
	confInfo := strings.Join(fields, "\n")
	log.Println("config of this simulation is\n" + confInfo)

	f, err := os.OpenFile("config.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err2 := f.WriteString(confInfo)

	if err2 != nil {
		log.Fatal(err2)
	}

}
