package tts

import (
	"errors"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/wujunwei928/edge-tts-go/edge_tts"
	"go.uber.org/zap"
	"os"
	"virtugo/logs"
)

var (
	audioData  []byte
	listVoices bool
	text       string
	file       string
	voice      string
	rate       string
	volume     string
	pitch      string
	wordsInCue float64
	writeMedia string
	proxyURL   string // 是否使用代理
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "edge-tts-go",
	Short:   "调用Edge TTS服务，文本生成语音",
	Long:    `调用Edge TTS服务，文本生成语音`,
	Version: edge_tts.PackageVersion, // 指定版本号: 会有 -v 和 --version 选项, 用于打印版本号
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		// 列出可用的语音

		// 文本转语音

		if len(text) <= 0 && len(file) <= 0 {
			return errors.New("--text and --file can't be empty at the same time")
		}

		inputText := text

		connOptions := []edge_tts.CommunicateOption{
			edge_tts.SetVoice(voice),
			edge_tts.SetRate(rate),
			edge_tts.SetVolume(volume),
			edge_tts.SetPitch(pitch),
			edge_tts.SetReceiveTimeout(20),
		}
		if len(proxyURL) > 0 {
			connOptions = append(connOptions, edge_tts.SetProxy(proxyURL))
		}

		conn, err := edge_tts.NewCommunicate(
			inputText,
			connOptions...,
		)
		if err != nil {
			return err
		}
		audioData, err = conn.Stream()
		if err != nil {
			return err
		}

		return nil
	},
}

type EdgeTTS struct {
	voice string
}

func NewEdgeTTS(voice string) *EdgeTTS {
	if voice == "" {
		voice = "zh-CN-XiaoyiNeural"
	}
	return &EdgeTTS{voice: voice}
}

func (e *EdgeTTS) Generate(text string) (string, error) {

	// 直接使用edge-tts-go库API，不通过命令行
	conn, err := edge_tts.NewCommunicate(
		text,
		edge_tts.SetVoice(e.voice),
		edge_tts.SetRate("+0%"),
		edge_tts.SetVolume("+0%"),
		edge_tts.SetPitch("+0Hz"),
		edge_tts.SetReceiveTimeout(20),
	)
	if err != nil {
		logs.Logger.Error("创建EdgeTTS连接失败", zap.Error(err))
		return "", err
	}

	audioData, err := conn.Stream()
	if err != nil {
		logs.Logger.Error("TTS执行失败", zap.Error(err))
		return "", nil
	}
	// 保存 MP3 文件
	err = os.MkdirAll("cache", 0755) // Create the directory if it doesn't exist
	if err != nil {
		logs.Logger.Error("创建cache目录失败", zap.Error(err))
		return "", err
	}
	filename := uuid.New().String()
	path := "cache/" + filename + ".mp3"
	writeMediaErr := os.WriteFile(path, audioData, 0644)
	if writeMediaErr != nil {
		logs.Logger.Error("写入MP3文件失败", zap.Error(writeMediaErr))
		return path, err
	}
	return path, nil
}
