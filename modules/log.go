// modules/log.go
package modules

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

const (
	DEBUG   = "DEBUG"
	INFO    = "INFO"
	WARNING = "WARNING"
	ERROR   = "ERROR"
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
	logger = log.New(multiWriter, "", 0)
	// logger = log.New(multiWriter, "", log.Ldate|log.Ltime|log.Lshortfile)

	// 记录初始化日志
	Log("Log initialized successfully.", INFO)
	return nil
}

// 关闭日志文件
func CloseLog() {
	if logFile != nil {
		logFile.Close()
		Log("Log file closed.", INFO)
	}
}

// 提供一个用于日志记录的函数
func Log(message string, level string) {
	// 获取当前时间
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	//根据日志等级，设置日志前缀
	var logMessage string
	switch level {
	case DEBUG:
		logMessage = fmt.Sprintf("[%s] [DEBUG] %s: %s", timestamp, level, message)
	case INFO:
		logMessage = fmt.Sprintf("[%s] [INFO] %s: %s", timestamp, level, message)
	case WARNING:
		logMessage = fmt.Sprintf("[%s] [WARNING] %s: %s", timestamp, level, message)
	case ERROR:
		logMessage = fmt.Sprintf("[%s] [ERROR] %s: %s", timestamp, level, message)
	default:
		logMessage = fmt.Sprintf("[%s] [INFO] %s: %s", timestamp, INFO, message)
	}

	// 输出日志到控制台和文件
	logger.Println(logMessage)
}

func Log_INFO(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] [INFO] %s: %s", timestamp, INFO, message)
	logger.Println(logMessage)
}
