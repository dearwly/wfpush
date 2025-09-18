package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	"wfpush/modules"
)

func main() {
	// 如果用户输入了命令行参数，则执行指令模式
	if len(os.Args) > 1 {
		handleCommand()
		return // 执行完指令后退出
	}

	// 如果没有参数，则启动裂缝监控服务模式
	runService()
}

// handleCommand 负责处理 list, add, delete 等指令
func handleCommand() {
	command := os.Args[1]
	args := os.Args[2:]

	// --- 核心修改部分 ---
	// 我们现在捕获指令返回的字符串，并将其打印出来
	switch command {
	case "list":
		result, _ := modules.ListSubscriptions()
		fmt.Println(result)
	case "add":
		result, _ := modules.AddSubscription(args)
		fmt.Println(result)
	case "delete":
		if len(args) != 1 {
			fmt.Println("用法: delete <索引>")
			fmt.Println("使用 'list' 命令查看订阅及其索引。")
			return
		}
		result, _ := modules.DeleteSubscription(args[0])
		fmt.Println(result)
	case "help":
		printHelp()
	default:
		fmt.Printf("未知指令: '%s'\n", command)
		printHelp()
	}
	// --- 修改结束 ---
}

// runService 启动持续监控裂缝的服务
func runService() {
	fmt.Println("启动裂缝订阅监控服务...")
	cfg, err := modules.LoadConfig(modules.GetAbsPath("config.yml"))
	if err != nil {
		fmt.Println("错误:", err)
		// 如果是初次创建文件导致的错误，提示后直接退出
		if strings.Contains(err.Error(), "请先填写") {
			return
		}
	}

	if err := initialData(); err != nil {
		// Log 函数此时可能还未初始化，所以用 fmt
		fmt.Println("初始化失败:", err)
		return
	}
	defer modules.CloseLog()

	if cfg.QQ.Enabled {
		go modules.StartQQBotServer(cfg)
	}

	warframe := modules.Warframe{
		SubsFissures: cfg.Subscriptions,
	}

	modules.Log("正在执行首次裂缝检查...", modules.INFO)
	err = modules.SendSubsFissures(warframe, cfg)
	if err != nil {
		modules.Log(fmt.Sprint("首次检查出错: ", err), modules.ERROR)
	}

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	modules.Log("服务已启动，每5分钟检查一次裂缝更新。", modules.INFO)
	for range ticker.C {
		err = modules.SendSubsFissures(warframe, cfg)
		if err != nil {
			modules.Log(fmt.Sprint("执行轮询任务出错: ", err), modules.ERROR)
		}
	}
}

// printHelp 显示帮助信息
func printHelp() {
	fmt.Println("\nWarframe 裂缝订阅助手")
	fmt.Println("用法:")
	fmt.Println("  <无参数>         - 启动裂缝监控服务")
	fmt.Println("  list             - 列出所有当前的订阅")
	fmt.Println("  add <任务> [...] - 添加一个新的订阅")
	fmt.Println("  delete <索引>    - 根据索引删除一个订阅")
	fmt.Println("  help             - 显示此帮助信息")
	fmt.Println("\n'add' 命令示例:")
	fmt.Println("  add 捕获")
	fmt.Println("  add 生存 is_hard=true")
	fmt.Println("  add 挖掘 tier=古纪 is_hard=false")
}

// initialData 函数保持不变
func initialData() error {
	err := modules.InitLog(modules.GetAbsPath("log.txt"))
	if err != nil {
		return err
	}
	if _, err := os.Stat(modules.GetAbsPath("data.json")); os.IsNotExist(err) {
		initialData := map[string][]modules.Fissure{
			"fissure": {},
		}
		file, err := os.Create(modules.GetAbsPath("data.json"))
		if err != nil {
			return err
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(initialData)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
