package main

import (
	"github.com/spf13/viper"
)

// Config 全局的配置,在main开始前初始化
var Config struct {
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
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	if err := viper.ReadInConfig(); err != nil {
		panic("config file not find")
	}

	Config.NodeNum = int32(viper.GetInt("NodeNum"))
	Config.NetLatency = int32(viper.GetInt("NetLatency"))
	Config.DcssNeibor = int32(viper.GetInt("DcssNeibor"))
	Config.NodeCpu = int32(viper.GetInt("NodeCpu"))
	Config.NodeMemory = int32(viper.GetInt("NodeMemory"))
	Config.TaskNumFactor = float32(viper.GetFloat64("TaskNumFactor"))
	Config.TaskCpu = int32(viper.GetInt("TaskCpu"))
	Config.TaskMemory = int32(viper.GetInt("TaskMemory"))
	Config.TaskLifeTime = int32(viper.GetInt("TaskLifeTime"))
	Config.SchedulerPerformance = float32(viper.GetInt("SchedulerPerformance"))
	Config.StateUpdatePeriod = int32(viper.GetInt("StateUpdatePeriod"))

}
