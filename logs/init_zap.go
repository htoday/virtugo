package logs

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

var Logger *zap.Logger

func InitZap() {
	// 创建日志编码器配置
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000")) // 自定义时间格式
	}
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 启用彩色日志级别编码器

	// 创建控制台编码器
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	// 创建日志核心，将日志输出到控制台
	core := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)

	// 创建 Logger 实例
	Logger = zap.New(core)
	defer Logger.Sync()
	Logger.Info("日志初始化成功")
}
