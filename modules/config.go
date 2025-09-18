package modules

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config 结构体对应 config.yml 的结构
type Config struct {
	Email struct {
		Enabled    bool     `yaml:"enabled"`
		SMTPServer string   `yaml:"smtp_server"`
		Port       int      `yaml:"port"`
		Sender     string   `yaml:"sender"`
		Password   string   `yaml:"password"`
		Recipients []string `yaml:"recipients"`
	} `yaml:"email"`
	Subscriptions []SubFissure `yaml:"subscriptions"`
}

// LoadConfig 从指定路径加载并解析 config.yml 文件
func LoadConfig(path string) (*Config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
