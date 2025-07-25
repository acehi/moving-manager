package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config 应用配置结构
type Config struct {
}

// 全局配置实例
var AppConfig Config

// InitConfig 初始化配置
func InitConfig(configPath string) error {
	// 读取配置文件
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	// 解析YAML配置
	if err := yaml.Unmarshal(data, &AppConfig); err != nil {
		return err
	}

	return nil
}