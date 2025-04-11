package main

import (
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
	"virtugo/internal/config"
	"virtugo/internal/dao"
	"virtugo/internal/sever"
	"virtugo/logs"
)

func main() {
	logs.InitZap()
	// 获取当前可执行文件（main.exe）的绝对路径
	exePath, err := os.Executable()
	if err != nil {
		logs.Logger.Panic("❌ 获取可执行文件路径失败:", zap.Error(err))
	}
	logs.Logger.Debug("exepath是:" + exePath)
	// 获取 main.exe 所在目录
	exeDir := filepath.Dir(exePath)
	// 处理 GoLand 临时编译路径问题
	if strings.Contains(exePath, "__go_build") {
		// 如果是 GoLand 临时编译环境，使用当前工作目录
		workDir, _ := os.Getwd()
		exeDir = workDir
		config.ModelDirRoot = "./"
	} else {
		config.ModelDirRoot = exeDir
	}

	// 计算 asrModel/model 目录
	modelPath := filepath.Join(exeDir, "asrModel", "model.int8.onnx")
	logs.Logger.Info("可执行文件目录:" + exeDir)
	logs.Logger.Info("模型路径:" + modelPath)
	// 检查文件是否存在
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		logs.Logger.Error("❌ 模型文件不存在" + modelPath)
	} else {
		logs.Logger.Info("✅ 模型文件存在，可以使用")
	}

	dao.InitChromemDB()
	dao.InitSqlite()
	config.LoadConfig(exeDir)
	config.LoadMCPJson()
	sever.StartSever("127.0.0.1", "8080")
}
