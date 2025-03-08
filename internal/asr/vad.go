package asr

import sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"

func InitVad() sherpa.VadModelConfig {
	// 1. Create VAD
	config := sherpa.VadModelConfig{}
	config.SileroVad.Model = "./asrModel/silero_vad.onnx"
	config.SileroVad.Threshold = 0.5
	config.SileroVad.MinSilenceDuration = 0.5
	config.SileroVad.MinSpeechDuration = 0.1
	config.SileroVad.WindowSize = 512
	config.SileroVad.MaxSpeechDuration = 10.0
	config.SampleRate = 16000
	config.NumThreads = 1
	config.Provider = "cpu"
	config.Debug = 1

	return config

	//defer sherpa.DeleteVoiceActivityDetector(vad)
}
