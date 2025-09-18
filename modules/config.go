package modules

import (
	"fmt"
	"io/ioutil"
	"os"

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

	// ---- QQ 配置结构体 ----
	QQ struct {
		Enabled         bool `yaml:"enabled"`
		WebSocketServer struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		} `yaml:"websocket_server"`
		PushTargets []struct {
			Type string `yaml:"type"`
			ID   string `yaml:"id"`
		} `yaml:"push_targets"`
	} `yaml:"qq"`

	Subscriptions []SubFissure `yaml:"subscriptions"`
}

// defaultConfigFileContent 包含了带有详细注释的默认配置
const defaultConfigFileContent = `# Warframe 裂缝订阅助手 - 配置文件
# ==========================================================

# 邮件通知功能配置
email:
  enabled: true
  smtp_server: "smtp.163.com"
  port: 465
  sender: "your_email@163.com"
  password: "your_app_password"
  recipients:
    - "recipient1@example.com"

# QQ 机器人推送配置
qq:
  # 是否启用 QQ 推送功能。
  # true: 启用 | false: 禁用
  enabled: true

  # WebSocket 服务器配置, 你的 QQ 机器人客户端将连接到这里。
  websocket_server:
    # 监听的 IP 地址。 "0.0.0.0" 表示监听所有网络接口。
    host: "0.0.0.0"
    # 监听的端口。确保这个端口没有被其他程序占用。
    port: 8088

  # 推送目标列表。你可以指定多个群组和/或个人。
  push_targets:
    # 示例 1: 推送到一个群聊
    - type: "group"
      id: 123456789   # 替换为你的目标群号

    # 示例 2: 推送到一个私聊
    - type: "private"
      id: 987654321   # 替换为你的目标 QQ 号

# 裂缝订阅列表
subscriptions:
  - mission_type: "歼灭"
    is_hard: false
  - mission_type: "捕获"
    is_hard: false
  - mission_type: "歼灭"
    is_hard: true
  - mission_type: "捕获"
    is_hard: true
`

// CreateDefaultConfig 创建一个带有注释的默认配置文件
func CreateDefaultConfig(path string) error {
	return ioutil.WriteFile(path, []byte(defaultConfigFileContent), 0644)
}

// LoadConfig 从指定路径加载并解析 config.yml 文件
func LoadConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("配置文件 '%s' 不存在，将为您创建一个新的。\n", path)
		if err := CreateDefaultConfig(path); err != nil {
			return nil, fmt.Errorf("创建默认配置文件失败: %w", err)
		}
		return nil, fmt.Errorf("请先填写新的配置文件 '%s' 后再重新运行程序", path)
	}

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

// SaveConfig 将配置结构体保存回 YAML 文件
func SaveConfig(path string, cfg *Config) error {
	bytes, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, bytes, 0644)
}
