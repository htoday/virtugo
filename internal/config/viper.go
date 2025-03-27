package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"log"
	"os"
	"virtugo/logs"
)

type Config struct {
	OpenAI struct {
		Key string `mapstructure:"key"`
	} `mapstructure:"openai_api_key"`
	Prompt struct {
		Persona string `mapstructure:"persona"`
	} `mapstructure:"prompt"`
	Temperature float32 `mapstructure:"Temperature"`
	TTS         struct {
		ServiceType    string `mapstructure:"service_type"`
		EdgeTTSVoice   string `mapstructure:"edge_tts_voice"`
		FishAudioKey   string `mapstructure:"fish_audio_key"`
		FishAudioVoice string `mapstructure:"fish_audio_voice"`
	} `mapstructure:"tts"`
}

var ModelDirRoot string
var Cfg Config
var Cwd string

// LoadConfig 读取 config.yaml，并自动设置环境变量
func LoadConfig(exeDir string) {
	// 获取可执行文件的目录

	// 让 viper 在可执行文件所在目录查找 config.yaml
	viper.SetConfigName("config") // 不要加 .yaml
	viper.SetConfigType("yaml")
	viper.AddConfigPath(exeDir) // 重点：让可执行文件和 config.yaml 放一起

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		//log.Printf("⚠️  未找到配置文件: %v", err)
		logs.Logger.Warn("⚠️  未找到配置文件:", zap.Error(err))
	}

	// 解析到结构体
	if err := viper.Unmarshal(&Cfg); err != nil {
		log.Fatalf("❌ 解析配置失败: %v", err)
	}

	// **优先检查是否已存在环境变量**
	if os.Getenv("OPENAI_API_KEY") == "" {
		if Cfg.OpenAI.Key != "" {
			_ = os.Setenv("OPENAI_API_KEY", Cfg.OpenAI.Key) // **自动设置环境变量**
			logs.Logger.Info("✅ 已从配置文件加载 OPENAI_API_KEY")
		} else {
			log.Fatalf("❌ 缺少 OPENAI_API_KEY，请设置环境变量或修改 config.yaml")
		}
	} else {
		logs.Logger.Info("✅ 发现已有环境变量 OPENAI_API_KEY")
	}
}

// GetConfig 获取当前配置
func GetConfig() *Config {
	return &Cfg
}
