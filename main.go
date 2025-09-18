package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
	"wfpush/modules"
)

func main() {
	err := initialData()
	if err != nil {
		modules.Log(fmt.Sprint("Error initializing :", err), modules.ERROR)
	}
	defer modules.CloseLog() // 确保程序退出时关闭日志文件
	// modules.Send_email()

	// 初始化一个 Warframe 结构体，并填充其中的 SubsFissures 切片
	warframe := modules.Warframe{
		SubsFissures: []modules.SubFissure{
			{
				MissionType: "殲滅",
			},
			{
				MissionType: "捕獲",
			},
			{
				MissionType: "殲滅",
				IsHard:      true,
			},
			{
				MissionType: "捕獲",
				IsHard:      true,
			},
		},
	}

	// 创建一个每5分钟触发一次的 Ticker
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// 使用 goroutine 或循环来不断等待 Ticker 信号并调用函数
	for range ticker.C { // 每当 ticker.C 触发时执行
		err = modules.SendSubsFissures(warframe)
		if err != nil {
			// 处理错误
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
	//初始化exists.json文件
	if _, err := os.Stat("exists.json"); os.IsNotExist(err) {
		// 如果文件不存在，创建并初始化
		initialData := map[string][]modules.Fissure{
			"fissure": {},
		}
		file, err := os.Create("exists.json")
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
