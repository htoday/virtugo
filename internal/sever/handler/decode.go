package handler

import (
	"bytes"
	"encoding/binary"
	"fmt"
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
	"log"
	"os"
	"os/exec"
	"strings"
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

func writeWAVHeader(file *os.File, sampleRate, numSamples int) {
	// RIFF 头
	file.Write([]byte("RIFF"))
	binary.Write(file, binary.LittleEndian, uint32(36+numSamples*2)) // 文件总长度
	file.Write([]byte("WAVE"))

	// fmt 子块
	file.Write([]byte("fmt "))
	binary.Write(file, binary.LittleEndian, uint32(16))           // fmt 块长度
	binary.Write(file, binary.LittleEndian, uint16(1))            // PCM 格式
	binary.Write(file, binary.LittleEndian, uint16(1))            // 单声道
	binary.Write(file, binary.LittleEndian, uint32(sampleRate))   // 采样率
	binary.Write(file, binary.LittleEndian, uint32(sampleRate*2)) // 字节率（采样率 * 通道数 * 位深/8）
	binary.Write(file, binary.LittleEndian, uint16(2))            // 块对齐（通道数 * 位深/8）
	binary.Write(file, binary.LittleEndian, uint16(16))           // 位深

	// data 子块
	file.Write([]byte("data"))
	binary.Write(file, binary.LittleEndian, uint32(numSamples*2)) // 数据长度
}

func decodeAudio(data []byte) []float32 {
	reader := bytes.NewReader(data)
	var pcmSamples []float32

	for {
		var sample int16
		err := binary.Read(reader, binary.LittleEndian, &sample)
		if err != nil {
			break
		}
		pcmSamples = append(pcmSamples, float32(sample)/32768.0) // 转换为浮点数
	}

	return pcmSamples
}

// 将字节数据转换为 Float32 数组
func byteToFloat32Array(data []byte) ([]float32, error) {
	var result []float32
	reader := bytes.NewReader(data)

	// 将每 4 个字节（即一个 float32）转换成一个 float32
	for {
		var sample float32
		err := binary.Read(reader, binary.LittleEndian, &sample) // 使用 LittleEndian 解析
		if err != nil {
			// 如果读到末尾，返回解析结果
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
		result = append(result, sample)
	}

	return result, nil
}
func convertWebMToPCM(webmData []byte) ([]byte, error) {
	cmd := exec.Command("ffmpeg", "-i", "pipe:0", "-f", "s16le", "-ac", "1", "-ar", "16000", "pipe:1")

	// 创建管道
	var stdoutBuf bytes.Buffer
	cmd.Stdin = bytes.NewReader(webmData)
	cmd.Stdout = &stdoutBuf

	// 运行 ffmpeg
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg 转换失败: %v", err)
	}

	// 返回 PCM 数据
	return stdoutBuf.Bytes(), nil
}

// **语音识别**
func recognizeSpeech(recognizer *sherpa.OfflineRecognizer, audio *sherpa.Wave) string {
	stream := sherpa.NewOfflineStream(recognizer)
	defer sherpa.DeleteOfflineStream(stream)
	stream.AcceptWaveform(audio.SampleRate, audio.Samples)

	recognizer.Decode(stream)
	result := stream.GetResult()
	if result == nil {
		return ""
	}
	log.Println("识别结果:", result.Text)
	return result.Text
}
func saveDebugWAV(data []float32) {
	file, _ := os.Create("debug.wav")
	defer file.Close()
	writeWAVHeader(file, 16000, len(data)*2)
	for _, v := range data {
		int16Val := int16(v * 32768)
		binary.Write(file, binary.LittleEndian, int16Val)
	}
}
func decode(recognizer *sherpa.OfflineRecognizer, audio *sherpa.Wave, id int) {
	stream := sherpa.NewOfflineStream(recognizer)
	defer sherpa.DeleteOfflineStream(stream)
	stream.AcceptWaveform(audio.SampleRate, audio.Samples)
	recognizer.Decode(stream)
	result := stream.GetResult()
	if result == nil {
		log.Println("No result")
		return
	}
	text := strings.ToLower(result.Text)
	text = strings.Trim(text, " ")
	log.Println(text)

	duration := float32(len(audio.Samples)) / float32(audio.SampleRate)

	filename := fmt.Sprintf("seg-%d-%.2f-seconds-%s.wav", id, duration, text)
	ok := audio.Save(filename)
	if ok {
		log.Printf("Saved to %s", filename)
	}
	log.Print("----------\n")
}
