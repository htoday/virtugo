package tts

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"virtugo/logs"
)

// ServeReferenceAudio 定义参考音频结构
type ServeReferenceAudio struct {
	Audio []byte `msgpack:"audio"`
	Text  string `msgpack:"text"`
}

// ServeTTSRequest 定义 TTS 请求结构
type ServeTTSRequest struct {
	Text        string                `msgpack:"text"`
	ChunkLength int                   `msgpack:"chunk_length"`
	Format      string                `msgpack:"format"`
	MP3Bitrate  int                   `msgpack:"mp3_bitrate"`
	References  []ServeReferenceAudio `msgpack:"references"`
	ReferenceID *string               `msgpack:"reference_id"`
	Normalize   bool                  `msgpack:"normalize"`
	Latency     string                `msgpack:"latency"`
}

func Use() {
	referenceID := "63fc280997324e0d8096c2e59d869f32"

	reqData := ServeTTSRequest{
		Text:        "",
		ChunkLength: 200,
		Format:      "mp3",
		MP3Bitrate:  128,
		References:  []ServeReferenceAudio{},
		ReferenceID: &referenceID,
		Normalize:   true,
		Latency:     "normal",
	}

	// 使用 msgpack 编码请求
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	err := enc.Encode(reqData)
	if err != nil {
		panic(err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", "https://api.fish.audio/v1/tts", &buf)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Bearer 525b512474b04f8cab927d0db13e9812") // 替换成你的 key
	req.Header.Set("Content-Type", "application/msgpack")
	req.Header.Set("Model", "speech-1.5") // 可选模型：speech-1.5 / 1.6 / agent-x0

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("API返回错误状态码: %d\n", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("错误详情: %s\n", string(body))
		return
	}
	// 创建输出音频文件
	outFile, err := os.Create("hello.mp3")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	// 逐块写入音频数据
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("音频已保存为 hello.mp3")
}

type FishAudioTTS struct {
	voice string
	key   string
}

func NewFishAudioTTS(voice string, key string) *FishAudioTTS {
	return &FishAudioTTS{
		voice: voice,
		key:   key,
	}
}

func (f *FishAudioTTS) Generate(text string) (string, error) {
	referenceID := f.voice

	reqData := ServeTTSRequest{
		Text:        text,
		ChunkLength: 200,
		Format:      "mp3",
		MP3Bitrate:  128,
		References:  []ServeReferenceAudio{},
		ReferenceID: &referenceID,
		Normalize:   true,
		Latency:     "normal",
	}

	// 使用 msgpack 编码请求
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	err := enc.Encode(reqData)
	if err != nil {
		return "", err
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", "https://api.fish.audio/v1/tts", &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+f.key) // 替换成你的 key
	req.Header.Set("Content-Type", "application/msgpack")
	req.Header.Set("Model", f.voice) // 可选模型：speech-1.5 / 1.6 / agent-x0

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API返回错误状态码: %d\n", resp.StatusCode)
	}

	err = os.MkdirAll("cache", 0755) // Create the directory if it doesn't exist
	if err != nil {
		logs.Logger.Error("创建cache目录失败", zap.Error(err))
	}
	// 创建输出音频文件
	filename := uuid.New().String() + ".mp3"
	path := "cache/" + filename
	outFile, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	// 逐块写入音频数据
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", err
	}
	return path, nil
}
