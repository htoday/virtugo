package config

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"log"
	"virtugo/logs"
)

type ModelInfo struct {
	ApiType      string  `mapstructure:"api_type"`
	RoleName     string  `mapstructure:"role_name"`
	ModelName    string  `mapstructure:"model_name"`
	BaseURL      string  `mapstructure:"base_url"`
	APIKey       string  `mapstructure:"api_key"`
	Temperature  float32 `mapstructure:"temperature"`
	Persona      string  `mapstructure:"persona"`
	SystemPrompt string  `mapstructure:"system_prompt"`
	Live2dModel  string  `mapstructure:"live2d_model"`
}

type TTSConfig struct {
	ServiceType    string `mapstructure:"service_type"`
	EdgeVoice      string `mapstructure:"edge_tts_voice"`
	FishAudioKey   string `mapstructure:"fish_audio_key"`
	FishAudioVoice string `mapstructure:"fish_audio_voice"`
}

type ModelConfig struct {
	ModelInfo ModelInfo   `mapstructure:"model_info"`
	TTS       TTSConfig   `mapstructure:"tts"`
	Tools     ToolsConfig `mapstructure:"tools"`
}
type ToolsConfig struct {
	DuckDuckGo struct {
		IsEnable bool `mapstructure:"is_enable"`
	} `mapstructure:"duckduckgo"`
	Wikipedia struct {
		IsEnable  bool   `mapstructure:"is_enable"`
		UserAgent string `mapstructure:"user_agent"`
	} `mapstructure:"wikipedia"`
	MCPTool struct {
		IsEnable bool `mapstructure:"is_enable"`
	} `mapstructure:"mcp_tool"`
}
type Config struct {
	Models            map[string]ModelConfig `mapstructure:"models"`
	PreGenerateAmount int                    `mapstructure:"pre_generate_amount"`
	//Language string                 `mapstructure:"language"`
	AuthKey         string `mapstructure:"auth_key"`
	KeyWordIsEnable bool   `mapstructure:"key_word_is_enable"` // 新增字段
	History         struct {
		MaxLength int `mapstructure:"max_length"`
	} `mapstructure:"history"`
	BackendPort  int `mapstructure:"backend_port"`
	FrontendPort int `mapstructure:"frontend_port"`
}

var ModelDirRoot string
var Cfg Config

// LoadConfig 读取 config.yaml，并自动设置环境变量
func LoadConfig(exeDir string) {
	// 获取可执行文件的目录

	// 让 viper 在可执行文件所在目录查找 config.yaml
	viper.SetConfigName("config") // 不要加 .yaml
	viper.SetConfigType("yaml")
	viper.AddConfigPath(exeDir) // 重点：让可执行文件和 config.yaml 放一起

	viper.SetDefault("pre_generate_amount", 1)    // 设置默认值
	viper.SetDefault("auth_key", "1145141919810") // 设置默认值
	viper.SetDefault("key_word_is_enable", false) // 设置默认值
	viper.SetDefault("backend_port", 8081)        // 设置默认值
	viper.SetDefault("frontend_port", 8082)       // 设置默认值
	viper.SetDefault("history.max_length", 10)    // 设置默认值

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		//log.Printf("⚠️  未找到配置文件: %v", err)
		logs.Logger.Warn("⚠️  未找到配置文件:", zap.Error(err))
	}

	// 解析到结构体
	if err := viper.Unmarshal(&Cfg); err != nil {
		log.Fatalf("❌ 解析配置失败: %v", err)
	}
	b, _ := json.MarshalIndent(Cfg, "", "  ")
	fmt.Println(string(b))
}
