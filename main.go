package main

import (
	"fmt"
	"time"
	"wfpush/modules"
)

func main() {
	// 初始化日志
	err := modules.InitLog("log.txt")
	if err != nil {
		fmt.Println("Error initializing log:", err)
		return
	}
	defer modules.CloseLog() // 确保程序退出时关闭日志文件

	// modules.Send_email()

	// 初始化一个 Warframe 结构体，并填充其中的 SubsFissures 切片
	warframe := modules.Warframe{
		SubsFissures: []modules.SubFissure{
			{
				MissionType: "中斷", IsHard: true,
			},
		},
	}

	// 创建一个每5分钟触发一次的 Ticker
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// 使用 goroutine 或循环来不断等待 Ticker 信号并调用函数
	for {
		select {
		case <-ticker.C: // 当 Ticker 发出信号时，执行函数
			err = modules.SendSubsFissures(warframe)
			if err != nil {
				// 处理错误
				modules.Log(fmt.Sprint("Error: ", err), modules.ERROR)
			}
		}
	}

}
