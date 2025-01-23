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

	// modules.Send_email()

	// 初始化一个 Warframe 结构体，并填充其中的 SubsFissures 切片
	warframe := modules.Warframe{
		SubsFissures: []modules.SubFissure{
			{
				MissionType: "殲滅",
			},
			{
				MissionType: "防禦",
			},
		},
	}

	// type_ := 0 // Get all types of fissures

	// fissures, err := modules.GetFissures(warframe, type_)
	// if err != nil {
	// 	modules.Log(fmt.Sprint("Error:", err), modules.ERROR)
	// 	return
	// }

	// // Output the fissures
	// for _, fissure := range fissures {
	// 	modules.Log_INFO(fmt.Sprint("ID: %s, Title: %s, Location: %s\n", fissure["id"], fissure["title"], fissure["location"]))
	// }

	err = modules.CheckSubFissure(warframe)
	if err != nil {
		modules.Log(fmt.Sprint("Error:", err), modules.ERROR)
		return
	}

}
