package conf

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

// Config 包含应用程序的所有配置信息
type Config struct {
	Header   map[string]string `yaml:"header"`
	LogLevel string            `yaml:"log_level"`
	Port     int               `yaml:"port"`
}

func LoadConfig(configFile string) (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

var Cfg *Config

func init() {
	cfgName := "config.yaml"
	config, err := LoadConfig("conf/" + cfgName)
	if err != nil {
		config, err = LoadConfig(cfgName)
		if err != nil {
			logrus.Warnf("failed to load config file: %v", err)
			//panic(err)
		}
	}
	Cfg = config
}
