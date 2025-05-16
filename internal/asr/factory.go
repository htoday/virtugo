package asr

import "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"

//其他方法需要把wave转换成对应的格式

type Recognizer interface {
	Recognize(audio *sherpa_onnx.Wave) (string, error)
}

func NewASR(mode string) Recognizer {
	switch mode {
	case "sherpa":
		return NewSherpaOfflineAsr()
	default:
		return NewSherpaOfflineAsr()
	}
}
