package asr

import (
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
	"path/filepath"
	"virtugo/internal/config"
)

func InitOfflineASR() *sherpa.OfflineRecognizer {
	var modelUrl = filepath.Join(config.ModelDirRoot, "asrModel", "model.int8.onnx")
	var tokenUrl = filepath.Join(config.ModelDirRoot, "asrModel", "tokens.txt")

	c := sherpa.OfflineRecognizerConfig{}
	c.FeatConfig.SampleRate = 16000
	c.FeatConfig.FeatureDim = 80

	// Please download the model from
	// https://github.com/k2-fsa/sherpa-onnx/releases/download/asr-models/sherpa-onnx-paraformer-trilingual-zh-cantonese-en.tar.bz2
	c.ModelConfig.SenseVoice.Model = modelUrl
	c.ModelConfig.SenseVoice.Language = "zh"
	c.ModelConfig.Tokens = tokenUrl

	//c.ModelConfig.Paraformer.Model = "./sherpa-onnx-paraformer-trilingual-zh-cantonese-en/model.int8.onnx"
	//c.ModelConfig.Tokens = "./sherpa-onnx-paraformer-trilingual-zh-cantonese-en/tokens.txt"
	c.ModelConfig.NumThreads = 2
	c.ModelConfig.Debug = 1
	c.ModelConfig.Provider = "cpu"

	return sherpa.NewOfflineRecognizer(&c)
}
