package conf

import "encoding/json"
import "io/ioutil"

type Config struct {
	ServerName string // 服务器名
	Port       string // 命令监听端口
	AdminPort  string // 管理员命令监听端口

	DbConfig    map[string]interface{} // 数据库配置信息
	RedisConfig map[string]interface{} // Redis配置信息
	LogConfig   map[string]interface{} // 日志配置
}

var globalConfig *Config

func Init(configFilePath string) (err error) {

	bytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return
	}

	globalConfig = new(Config)
	err = json.Unmarshal(bytes, globalConfig)
	if err != nil {
		globalConfig = nil
	}
	return
}

func GetConfig() *Config {
	return globalConfig
}
