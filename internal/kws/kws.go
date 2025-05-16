package kws

import (
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
	"path/filepath"
	config2 "virtugo/internal/config"
)

func InitKeyWordSpot() sherpa.KeywordSpotterConfig {
	EncoderUrl := filepath.Join(config2.ModelDirRoot, "kwsModel", "encoder-epoch-12-avg-2-chunk-16-left-64.onnx")
	DecoderUrl := filepath.Join(config2.ModelDirRoot, "kwsModel", "decoder-epoch-12-avg-2-chunk-16-left-64.onnx")
	JoinerUrl := filepath.Join(config2.ModelDirRoot, "kwsModel", "joiner-epoch-12-avg-2-chunk-16-left-64.onnx")
	TokensUrl := filepath.Join(config2.ModelDirRoot, "kwsModel", "tokens.txt")
	KeyWordUrl := filepath.Join(config2.ModelDirRoot, "kwsModel", "keywords.txt")
	config := sherpa.KeywordSpotterConfig{}

	config.ModelConfig.Transducer.Encoder = EncoderUrl
	config.ModelConfig.Transducer.Decoder = DecoderUrl
	config.ModelConfig.Transducer.Joiner = JoinerUrl
	config.ModelConfig.Tokens = TokensUrl
	config.KeywordsFile = KeyWordUrl
	config.ModelConfig.NumThreads = 1
	config.ModelConfig.Debug = 0

	return config
}

func NewKeyWordSpotter() *sherpa.KeywordSpotter {
	kwsConfig := InitKeyWordSpot()
	spotter := sherpa.NewKeywordSpotter(&kwsConfig)
	return spotter
}
