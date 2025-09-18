package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
	"wfpush/modules"
)

func main() {
	// 首先加载配置
	cfg, err := modules.LoadConfig("config.yml")
	if err != nil {
		// 如果配置文件加载失败，无法继续，直接打印错误并退出
		fmt.Println("Error loading config.yml:", err)
		return
	}

	err = initialData()
	if err != nil {
		modules.Log(fmt.Sprint("Error initializing :", err), modules.ERROR)
		return // 初始化失败也应退出
	}
	defer modules.CloseLog()

	// 使用从 config.yml 加载的订阅
	warframe := modules.Warframe{
		SubsFissures: cfg.Subscriptions,
	}

	// 首次运行立即执行一次检查
	modules.Log("Performing initial check...", modules.INFO)
	err = modules.SendSubsFissures(warframe, cfg)
	if err != nil {
		modules.Log(fmt.Sprint("Error during initial check: ", err), modules.ERROR)
	}

	// 创建一个每5分钟触发一次的 Ticker
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// 使用 goroutine 或循环来不断等待 Ticker 信号并调用函数
	for range ticker.C {
		// 将配置 cfg 传递给函数
		err = modules.SendSubsFissures(warframe, cfg)
		if err != nil {
			modules.Log(fmt.Sprint("Error: ", err), modules.ERROR)
		}
	}
}

func initialData() error {
	// 初始化日志
	err := modules.InitLog("log.txt")
	if err != nil {
		return err
	}
	//初始化data.json文件
	if _, err := os.Stat("data.json"); os.IsNotExist(err) {
		// 如果文件不存在，创建并初始化
		initialData := map[string][]modules.Fissure{
			"fissure": {},
		}
		file, err := os.Create("data.json")
		if err != nil {
			return err
		}
		defer file.Close()

		// 将初始数据写入文件
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
