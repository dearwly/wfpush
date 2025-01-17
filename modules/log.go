// modules/log.go
package modules

import (
	"fmt"
	"io"
	"log"
	"os"
)

// 定义一个全局日志文件指针
var logFile *os.File
var logger *log.Logger

// 初始化日志记录
func InitLog(logFileName string) error {
	var err error
	// 创建日志文件，如果文件不存在则创建，文件存在则追加
	logFile, err = os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("Failed to open log file: %v", err)
	}

	// 创建一个多路输出的 logger，将日志同时输出到文件和控制台
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger = log.New(multiWriter, "", log.Ldate|log.Ltime)

	// 记录初始化日志
	logger.Println("Log initialized successfully.")
	return nil
}

// 关闭日志文件
func CloseLog() {
	if logFile != nil {
		logFile.Close()
		logger.Println("Log file closed.")
	}
}

// 提供一个用于日志记录的函数
func Log(message string) {
	logger.Println(message)
}
