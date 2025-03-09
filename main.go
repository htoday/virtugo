package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"virtugo/internal/config"
	"virtugo/internal/dao"
	"virtugo/internal/sever"
)

func main() {
	// 获取当前可执行文件（main.exe）的绝对路径
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("❌ 获取可执行文件路径失败: %v", err)
	}
	log.Println("exepath", exePath)
	// 获取 main.exe 所在目录
	exeDir := filepath.Dir(exePath)
	// 处理 GoLand 临时编译路径问题
	if strings.Contains(exeDir, "JetBrains/GoLand") {
		// 如果是 GoLand 临时编译环境，使用当前工作目录
		workDir, _ := os.Getwd()
		exeDir = workDir
	}

	// 计算 asrModel/model 目录
	modelPath := filepath.Join(exeDir, "asrModel", "model.int8.onnx")

	fmt.Println("可执行文件目录:", exeDir)
	fmt.Println("模型路径:", modelPath)

	// 检查文件是否存在
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		log.Fatalf("❌ 模型文件不存在: %s", modelPath)
	}

	fmt.Println("✅ 模型文件存在，可以使用")
	dao.InitChromemDB()
	dao.InitSqlite()
	config.LoadConfig()
	sever.StartSever("127.0.0.1", "8080")
}
