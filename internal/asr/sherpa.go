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
	c.ModelConfig.Debug = 0
	c.ModelConfig.Provider = "cpu"

	return sherpa.NewOfflineRecognizer(&c)
}

type SherpaAsr struct {
	*sherpa.OfflineRecognizer
}

func NewSherpaOfflineAsr() SherpaAsr {
	return SherpaAsr{InitOfflineASR()}
}

func (s SherpaAsr) Recognize(audio *sherpa.Wave) (string, error) {
	stream := sherpa.NewOfflineStream(s.OfflineRecognizer)
	defer sherpa.DeleteOfflineStream(stream)
	stream.AcceptWaveform(audio.SampleRate, audio.Samples)
	s.Decode(stream)
	result := stream.GetResult()
	if result == nil {
		return "", nil
	}
	return result.Text, nil
}
