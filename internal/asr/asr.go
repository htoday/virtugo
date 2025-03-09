package asr

import (
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
	flag "github.com/spf13/pflag"
	"log"
	"path/filepath"
	"virtugo/internal/config"
)

var AscRec *sherpa.OnlineRecognizer

func InitOnlineAsr() {
	var modelUrl = filepath.Join(config.ModelDirRoot, "asrModel", "model.int8.onnx")
	//var tokenUrl = filepath.Join(config.ModelDirRoot, "asrModel", "tokens.txt")

	config := sherpa.OnlineRecognizerConfig{}
	config.FeatConfig = sherpa.FeatureConfig{SampleRate: 16000, FeatureDim: 80}

	flag.StringVar(&config.ModelConfig.Transducer.Encoder, "encoder", "", modelUrl)
	flag.StringVar(&config.ModelConfig.Transducer.Decoder, "decoder", "", modelUrl)
	flag.StringVar(&config.ModelConfig.Transducer.Joiner, "joiner", "", modelUrl)
	flag.StringVar(&config.ModelConfig.Paraformer.Encoder, "paraformer-encoder", "", modelUrl)
	flag.StringVar(&config.ModelConfig.Paraformer.Decoder, "paraformer-decoder", "", modelUrl)
	flag.StringVar(&config.ModelConfig.Tokens, "tokens", "", "Path to the tokens file")
	flag.IntVar(&config.ModelConfig.NumThreads, "num-threads", 1, "Number of threads for computing")
	flag.IntVar(&config.ModelConfig.Debug, "debug", 0, "Whether to show debug message")
	flag.StringVar(&config.ModelConfig.ModelType, "model-type", "", "Optional. Used for loading the model in a faster way")
	flag.StringVar(&config.ModelConfig.Provider, "provider", "cpu", "Provider to use")
	flag.StringVar(&config.DecodingMethod, "decoding-method", "greedy_search", "Decoding method. Possible values: greedy_search, modified_beam_search")
	flag.IntVar(&config.MaxActivePaths, "max-active-paths", 4, "Used only when --decoding-method is modified_beam_search")
	flag.IntVar(&config.EnableEndpoint, "enable-endpoint", 1, "Whether to enable endpoint")
	flag.Float32Var(&config.Rule1MinTrailingSilence, "rule1-min-trailing-silence", 2.4, "Threshold for rule1")
	flag.Float32Var(&config.Rule2MinTrailingSilence, "rule2-min-trailing-silence", 1.2, "Threshold for rule2")
	flag.Float32Var(&config.Rule3MinUtteranceLength, "rule3-min-utterance-length", 20, "Threshold for rule3")

	flag.Parse()

	log.Println("Initializing recognizer (may take several seconds)")
	AscRec = sherpa.NewOnlineRecognizer(&config)
	log.Println("Recognizer created!")
	//defer sherpa.DeleteOnlineRecognizer(AscRec)
}
