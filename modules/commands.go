package modules

import (
	"fmt"
	"strconv"
	"strings"
)

// ListSubscriptions 返回一个包含所有订阅的可读字符串
func ListSubscriptions() (string, error) {
	configPath := GetAbsPath("config.yml")
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Sprintf("错误：无法加载配置文件 '%s': %v", configPath, err), err
	}

	if len(cfg.Subscriptions) == 0 {
		return "当前没有任何订阅。\n使用 'add <任务类型> [is_hard=true/false] [tier=纪元]' 来添加。", nil
	}

	var builder strings.Builder
	builder.WriteString("--- 当前的裂缝订阅 ---\n")
	for i, sub := range cfg.Subscriptions {
		var parts []string
		parts = append(parts, sub.MissionType)
		if sub.IsHard {
			parts = append(parts, "(钢铁之路)")
		}
		if sub.Tier != "" {
			parts = append(parts, fmt.Sprintf("(纪元: %s)", sub.Tier))
		}
		builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, strings.Join(parts, " ")))
	}
	builder.WriteString("--------------------")
	return builder.String(), nil
}

// AddSubscription 添加新订阅并返回结果字符串
func AddSubscription(args []string) (string, error) {
	if len(args) < 1 {
		return "错误：请提供至少一个任务类型。\n用法: add <任务类型> [is_hard=true/false] [tier=纪元]\n示例: add 生存 is_hard=true tier=后纪", fmt.Errorf("参数不足")
	}

	newSub := SubFissure{
		MissionType: args[0],
	}

	for _, arg := range args[1:] {
		parts := strings.Split(arg, "=")
		if len(parts) != 2 {
			continue // 忽略无效参数
		}
		key := strings.ToLower(parts[0])
		value := parts[1]

		switch key {
		case "is_hard":
			isHard, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Sprintf("错误：'is_hard' 的值无效 '%s'，必须是 'true' 或 'false'。", value), err
			}
			newSub.IsHard = isHard
		case "tier":
			newSub.Tier = value
		}
	}

	configPath := GetAbsPath("config.yml")
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Sprintf("错误：无法加载配置文件 '%s': %v", configPath, err), err
	}

	cfg.Subscriptions = append(cfg.Subscriptions, newSub)

	if err := SaveConfig(configPath, cfg); err != nil {
		return fmt.Sprintf("错误：无法保存配置: %v", err), err
	}

	successMsg := fmt.Sprintf("✔ 成功添加订阅: %s", args[0])
	updatedList, _ := ListSubscriptions()
	return successMsg + "\n" + updatedList, nil
}

// DeleteSubscription 删除订阅并返回结果字符串
func DeleteSubscription(indexStr string) (string, error) {
	index, err := strconv.Atoi(indexStr)
	if err != nil || index <= 0 {
		return "错误：请输入一个有效的、大于0的数字索引。\n使用 'list' 命令查看订阅及其索引。", err
	}

	configPath := GetAbsPath("config.yml")
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Sprintf("错误：无法加载配置文件 '%s': %v", configPath, err), err
	}

	if index > len(cfg.Subscriptions) {
		return fmt.Sprintf("错误：索引 %d 超出范围。当前只有 %d 个订阅。", index, len(cfg.Subscriptions)), fmt.Errorf("索引越界")
	}

	removeIndex := index - 1
	removedSub := cfg.Subscriptions[removeIndex]
	cfg.Subscriptions = append(cfg.Subscriptions[:removeIndex], cfg.Subscriptions[removeIndex+1:]...)

	if err := SaveConfig(configPath, cfg); err != nil {
		return fmt.Sprintf("错误：无法保存配置: %v", err), err
	}

	successMsg := fmt.Sprintf("✔ 成功删除订阅: %s", removedSub.MissionType)
	updatedList, _ := ListSubscriptions()
	return successMsg + "\n" + updatedList, nil
}
