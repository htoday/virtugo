package asr

import (
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
	"path/filepath"
	config2 "virtugo/internal/config"
	"virtugo/logs"
)

func InitVad() sherpa.VadModelConfig {
	// 1. Create VAD
	//vadUrl := filepath.Join(config.Cwd, "asrModel", "tokens.txt")
	vadUrl := filepath.Join(config2.ModelDirRoot, "asrModel", "silero_vad.onnx")
	logs.Logger.Debug("vadUrl是" + vadUrl)
	config := sherpa.VadModelConfig{}
	config.SileroVad.Model = vadUrl
	config.SileroVad.Threshold = 0.5
	config.SileroVad.MinSilenceDuration = 0.5
	config.SileroVad.MinSpeechDuration = 0.1
	config.SileroVad.WindowSize = 512
	config.SileroVad.MaxSpeechDuration = 10.0
	config.SampleRate = 16000
	config.NumThreads = 1
	config.Provider = "cpu"
	config.Debug = 0

	return config

	//defer sherpa.DeleteVoiceActivityDetector(vad)
}
