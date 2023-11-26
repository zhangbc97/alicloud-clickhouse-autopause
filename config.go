package main

import "github.com/spf13/viper"

type ServerConfig struct {
	Port int `json:"port"`

	IdleSecondsBeforeStop     int          `json:"idleSecondsBeforeStop"`
	IdleCheckIntervalSeconds  int          `json:"idleCheckIntervalSeconds"`
	SyncStatusIntervalSeconds int          `json:"syncStatusIntervalSeconds"`
	WaitStatusIntervalSeconds int          `json:"waitStatusIntervalSeconds"`
	DbInstances               []DBInstance `json:"dbInstances"`
}
type DBInstance struct {
	RegionID        string `json:"regionID"`
	DBInstanceID    string `json:"dbInstanceID"`
	AccessKeyID     string `json:"accessKeyID"`
	AccessKeySecret string `json:"accessKeySecret"`
	ServerAddress   string `json:"serverAddress"`
	ServerHTTPPort  int    `json:"serverHttpPort"`
	ServerTCPPort   int    `json:"serverTcpPort"`
}

// ReadConfig 给定文件路径，从YAML中读取配置，返回配置结构体
func ReadConfig(path string) (ServerConfig, error) {
	// 使用Viper读取配置文件
	viper.SetConfigFile(path)
	err := viper.ReadInConfig()
	if err != nil {
		return ServerConfig{}, err
	}

	// 将配置文件中的配置读取到结构体中
	var config ServerConfig
	err = viper.Unmarshal(&config)
	if err != nil {
		return ServerConfig{}, err
	}

	return config, nil
}
