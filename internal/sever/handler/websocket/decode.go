package websocket

import (
	"bytes"
	"encoding/binary"
	"log"
)

func samplesInt16ToFloat(inSamples []byte) []float32 {
	numSamples := len(inSamples) / 2
	if len(inSamples)%2 != 0 {
		log.Println("Received odd-length data, truncating last byte")
		inSamples = inSamples[:len(inSamples)-1] // 丢弃最后一个字节
	}

	outSamples := make([]float32, numSamples)
	buf := bytes.NewReader(inSamples)

	for i := 0; i < numSamples; i++ {
		var sample int16
		err := binary.Read(buf, binary.LittleEndian, &sample)
		if err != nil {
			log.Fatal("Failed to parse 16-bit sample:", err)
		}
		outSamples[i] = float32(sample) / 32768.0
	}

	return outSamples
}

// **语音识别**
//func recognizeSpeech(recognizer *sherpa.OfflineRecognizer, audio *sherpa.Wave) string {
//	stream := sherpa.NewOfflineStream(recognizer)
//	defer sherpa.DeleteOfflineStream(stream)
//	stream.AcceptWaveform(audio.SampleRate, audio.Samples)
//	recognizer.Decode(stream)
//	result := stream.GetResult()
//	if result == nil {
//		return ""
//	}
//	logs.Logger.Info("识别到文本：" + result.Text)
//	return result.Text
//}
