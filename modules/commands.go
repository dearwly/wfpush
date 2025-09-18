package modules

import (
	"fmt"
	"strconv"
	"strings"
)

const configPath = "config.yml"

// ListSubscriptions 列出当前配置文件中的所有订阅
func ListSubscriptions() {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		fmt.Printf("错误：无法加载配置文件 '%s': %v\n", configPath, err)
		return
	}

	fmt.Println("--- 当前的裂缝订阅 ---")
	if len(cfg.Subscriptions) == 0 {
		fmt.Println("当前没有任何订阅。")
		fmt.Println("使用 'add <任务类型> [is_hard=true/false] [tier=纪元]' 来添加一个新订阅。")
		return
	}

	for i, sub := range cfg.Subscriptions {
		// 构建一个可读的描述
		var parts []string
		parts = append(parts, sub.MissionType)
		if sub.IsHard {
			parts = append(parts, "(钢铁之路)")
		}
		if sub.Tier != "" {
			parts = append(parts, fmt.Sprintf("(纪元: %s)", sub.Tier))
		}
		fmt.Printf("%d. %s\n", i+1, strings.Join(parts, " "))
	}
	fmt.Println("--------------------")
}

// AddSubscription 添加一个新的订阅并保存到配置文件
func AddSubscription(args []string) {
	if len(args) < 1 {
		fmt.Println("错误：请提供至少一个任务类型。")
		fmt.Println("用法: add <任务类型> [is_hard=true/false] [tier=纪元]")
		fmt.Println("示例: add 生存 is_hard=true tier=后纪")
		return
	}

	newSub := SubFissure{
		MissionType: args[0], // 第一个参数总是任务类型
	}

	// 解析可选参数 (is_hard, tier)
	for _, arg := range args[1:] {
		parts := strings.Split(arg, "=")
		if len(parts) != 2 {
			fmt.Printf("警告：无法解析参数 '%s'，已跳过。\n", arg)
			continue
		}
		key := strings.ToLower(parts[0])
		value := parts[1]

		switch key {
		case "is_hard":
			isHard, err := strconv.ParseBool(value)
			if err != nil {
				fmt.Printf("错误：'is_hard' 的值无效 '%s'，必须是 'true' 或 'false'。\n", value)
				return
			}
			newSub.IsHard = isHard
		case "tier":
			newSub.Tier = value
		default:
			fmt.Printf("警告：未知参数 '%s'，已跳过。\n", key)
		}
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		fmt.Printf("错误：无法加载配置文件 '%s': %v\n", configPath, err)
		return
	}

	cfg.Subscriptions = append(cfg.Subscriptions, newSub)

	if err := SaveConfig(configPath, cfg); err != nil {
		fmt.Printf("错误：无法保存配置: %v\n", err)
		return
	}

	fmt.Printf("✔ 成功添加订阅: %s\n", args[0])
	ListSubscriptions() // 显示更新后的列表
}

// DeleteSubscription 根据索引删除一个订阅
func DeleteSubscription(indexStr string) {
	index, err := strconv.Atoi(indexStr)
	if err != nil || index <= 0 {
		fmt.Println("错误：请输入一个有效的、大于0的数字索引。")
		fmt.Println("使用 'list' 命令查看订阅及其索引。")
		return
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		fmt.Printf("错误：无法加载配置文件 '%s': %v\n", configPath, err)
		return
	}

	if index > len(cfg.Subscriptions) {
		fmt.Printf("错误：索引 %d 超出范围。当前只有 %d 个订阅。\n", index, len(cfg.Subscriptions))
		return
	}

	// Go 中从切片中删除元素的常用方法
	// 将用户输入的 1-based 索引转为 0-based
	removeIndex := index - 1
	removedSub := cfg.Subscriptions[removeIndex]
	cfg.Subscriptions = append(cfg.Subscriptions[:removeIndex], cfg.Subscriptions[removeIndex+1:]...)

	if err := SaveConfig(configPath, cfg); err != nil {
		fmt.Printf("错误：无法保存配置: %v\n", err)
		return
	}

	fmt.Printf("✔ 成功删除订阅: %s\n", removedSub.MissionType)
	ListSubscriptions() // 显示更新后的列表
}
