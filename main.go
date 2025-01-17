package main

import (
	"fmt"
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

	modules.Send_email()

}
