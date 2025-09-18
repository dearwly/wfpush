package modules

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// ... Config 结构体保持不变 ...
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

// defaultConfigFileContent 包含了带有详细注释的默认配置
const defaultConfigFileContent = `# Warframe 裂缝订阅助手 - 配置文件
# ==========================================================

# 邮件通知功能配置
email:
  # 是否启用邮件发送功能。设置为 false 可关闭邮件，只在控制台打印日志。
  # true: 启用 | false: 禁用
  enabled: true

  # 邮件服务器 (SMTP) 地址。
  # 常见邮箱的 SMTP 服务器地址:
  #   - 163   邮箱: smtp.163.com
  #   - QQ    邮箱: smtp.qq.com
  #   - Gmail 邮箱: smtp.gmail.com
  smtp_server: "smtp.163.com"

  # SMTP 服务器端口。通常，使用 SSL/TLS 加密的端口是 465。
  port: 465

  # 您的发件人邮箱地址。
  sender: "your_email@163.com"

  # 您的邮箱密码或 "应用专用授权码"。
  # !!! 安全提示 !!!
  # 为了安全，强烈建议使用邮箱服务商提供的 "应用专用授权码" 而不是您的邮箱主密码。
  # 例如，163邮箱需要登录网页版，在设置中开启SMTP服务并生成授权码。
  password: "your_app_password"

  # 收件人列表。您可以添加一个或多个邮箱地址。
  recipients:
    - "recipient1@example.com"
    - "recipient2@example.com"

# 裂缝订阅列表
# 在这里定义您感兴趣的裂缝类型。
# 当符合以下任一条件的 *新* 裂缝出现时，您会收到通知。
subscriptions:
  # 示例 1: 订阅所有普通的“歼灭”任务
  - mission_type: "歼灭"
    is_hard: false

  # 示例 2: 订阅所有普通的“捕获”任务
  - mission_type: "捕获"
    is_hard: false

  # 示例 3: 订阅所有“钢铁之路”的“歼灭”任务
  - mission_type: "歼灭"
    is_hard: true

  # 示例 4: 订阅所有“钢铁之路”的“生存”任务，并且纪元是“后纪”
  - mission_type: "生存"
    is_hard: true
    tier: "后纪"

  # 字段说明:
  #   mission_type (必需): 任务类型。例如: "歼灭", "捕获", "生存", "挖掘" 等。
  #   is_hard (可选, 默认为 false): 是否为钢铁之路任务。true 或 false。
  #   tier (可选, 默认为所有纪元): 遗物纪元。可选值: "古纪", "前纪", "中纪", "后纪", "安魂"。
`

// CreateDefaultConfig 创建一个带有注释的默认配置文件
func CreateDefaultConfig(path string) error {
	return ioutil.WriteFile(path, []byte(defaultConfigFileContent), 0644)
}

// LoadConfig 修改：增加自动创建功能
func LoadConfig(path string) (*Config, error) {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("配置文件 '%s' 不存在，将为您创建一个新的。\n", path)
		if err := CreateDefaultConfig(path); err != nil {
			return nil, fmt.Errorf("创建默认配置文件失败: %w", err)
		}
		// 提示用户并返回一个特定错误，让主程序知道需要退出
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
