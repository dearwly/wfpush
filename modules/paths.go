package modules

import (
	"os"
	"path/filepath"
)

// appRoot 存储了可执行文件所在的目录的绝对路径
var appRoot string

// init 函数在包被加载时自动执行，这是设置全局变量的理想位置。
func init() {
	// 获取当前可执行文件的完整路径（例如 /path/to/your/program）
	exePath, err := os.Executable()
	if err != nil {
		// 如果无法获取路径，这是一个严重问题，程序无法继续
		panic("无法确定可执行文件路径: " + err.Error())
	}
	// 从完整路径中提取目录部分（例如 /path/to/your）
	appRoot = filepath.Dir(exePath)
}

// GetAbsPath 接受一个相对于程序根目录的文件名，
// 并返回其在文件系统中的绝对路径。
// 例如: GetAbsPath("config.yml") -> "/path/to/your/config.yml"
func GetAbsPath(filename string) string {
	return filepath.Join(appRoot, filename)
}
