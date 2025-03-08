package handler

import (
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"virtugo/internal/asr"
	"virtugo/internal/sever/context"
	model "virtugo/internal/sever/message_model"
	"virtugo/internal/tts"
)

type SafeBuffer struct {
	buf  []float32
	lock sync.Mutex
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // 允许跨域
}

func HandleWebsocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket 升级失败:", err)
		return
	}
	defer conn.Close()
	serviceContext := context.ServiceContext{}
	serviceContext.InitServiceContext()

	recAsr := asr.InitOfflineASR()
	vadConfig := asr.InitVad()
	var bufferSizeInSeconds float32 = 20
	vad := sherpa.NewVoiceActivityDetector(&vadConfig, bufferSizeInSeconds)
	// 缓存音频
	//var audioBuffer SafeBuffer
	var mutex sync.Mutex
	//var lastSpeechTime time.Time
	var printed bool

	k := 0
	fmt.Println("创建连接成功")
	for {
		// 读取消息
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("读取消息错误:", err)
			break
		}

		switch messageType {
		case websocket.TextMessage: // 文本消息
			fmt.Println("收到文本消息:", string(msg))
			var inputMessage model.TextMessage
			err := json.Unmarshal(msg, &inputMessage)
			if err != nil {
				log.Println("解析消息错误:", err)
			}
			switch inputMessage.Type {
			case "text":
				//err := serviceContext.AddUserQuestion(inputMessage.Content)
				//if err != nil {
				//	log.Println("添加用户问题失败:", err)
				//}
				mutex.Lock()
				streamResult, err := serviceContext.Stream(inputMessage.Content)
				if err != nil {
					log.Println("获取流失败:", err)
					continue
				}
				reportStream(streamResult, conn)
				mutex.Unlock()
			case "image":
				fmt.Println("处理图片消息:", inputMessage.Content)
			default:
				fmt.Println("未知消息类型:", inputMessage.Type)
			}
			//conn.WriteMessage(websocket.TextMessage, []byte(response))

		case websocket.BinaryMessage: // 二进制消息
			//log.Println("收到二进制消息，长度:", len(msg))
			// 识别音频
			// 解析 PCM 数据
			pcmData := samplesInt16ToFloat(msg)

			vad.AcceptWaveform(pcmData)

			if vad.IsSpeech() && !printed {
				printed = true
				log.Print("Detected speech\n")
				interrupt_msg := map[string]interface{}{
					"type":    "interrupt",
					"content": "interrupt",
				}

				fmt.Println("发送中断消息")
				err = conn.WriteJSON(interrupt_msg)
				if err != nil {
					log.Println("发送中断消息失败:", err)
				}
			}

			if !vad.IsSpeech() {
				printed = false
			}
			for !vad.IsEmpty() {
				speechSegment := vad.Front()
				vad.Pop()

				audio := &sherpa.Wave{}
				audio.Samples = speechSegment.Samples
				audio.SampleRate = 16000

				// Now decode it
				//go decode(recAsr, audio, k)
				go func() {
					text := recognizeSpeech(recAsr, audio)
					mutex.Lock()
					streamResult, err := serviceContext.Stream(text)
					if err != nil {
						log.Println("获取流失败:", err)
					}
					reportStream(streamResult, conn)
					mutex.Unlock()
				}()

				k += 1
			}
			//log.Println("PCM Data Length:", len(pcmData))
			// 传入 VAD 进行语音检测
			//vad.AcceptWaveform(pcmData)
			//// 判断是否检测到语音
			//if !vad.IsSpeech() {
			//	fmt.Println("未检测到语音")
			//} else {
			//	fmt.Println("检测到语音")
			//}

			//if vad.IsSpeech() {
			//	lastSpeechTime = time.Now()
			//	audioBuffer.lock.Lock()
			//	audioBuffer.buf = append(audioBuffer.buf, pcmData...)
			//	audioBuffer.lock.Unlock()
			//	if !printed {
			//		printed = true
			//		log.Print("Detected speech\n")
			//	}
			//} else if printed && time.Since(lastSpeechTime) > 1*time.Second {
			//	log.Print("Speech finished, processing...\n")
			//	//speechSegment := vad.Front()
			//	//vad.Pop()
			//
			//	audio := &sherpa.Wave{}
			//	audioBuffer.lock.Lock()
			//	audio.Samples = audioBuffer.buf
			//	audioBuffer.lock.Unlock()
			//	audio.SampleRate = 16000
			//	go func() {
			//		result := recognizeSpeech(recAsr, audio)
			//		log.Println("ASR 结果:", result)
			//
			//	}()
			//	audioBuffer.lock.Lock()
			//	audioBuffer.buf = nil
			//	audioBuffer.lock.Unlock()
			//	printed = false
			//}
			//if vad.IsSpeech() {
			//	if !isRecording {
			//		log.Println("Detected Speech, Start Recording...")
			//		isRecording = true
			//	}

			// 加锁缓存音频
			//mutex.Lock()
			//audioBuffer = append(audioBuffer, pcmData...)
			//mutex.Unlock()
			//} else if isRecording {
			//	// 语音结束，开始 ASR
			//	isRecording = false
			//	log.Println("发言结束，开始asr。。。")
			//
			//	// 复制缓存音频进行处理
			//	mutex.Lock()
			//	audioSegment := make([]float32, len(audioBuffer))
			//	copy(audioSegment, audioBuffer)
			//	audioBuffer = nil // 清空缓存
			//	mutex.Unlock()
			//	go func(data []float32) {
			//		wave := &sherpa.Wave{Samples: data, SampleRate: 16000}
			//		result := recognizeSpeech(recAsr, wave)
			//
			//		log.Println("asr的结果:", result)
			//		//if result != "" {
			//		//	streamResult, err := serviceContext.Stream(result)
			//		//	if err != nil {
			//		//		log.Println("获取流失败:", err)
			//		//	}
			//		//	reportStream(streamResult, conn)
			//		//}
			//	}(audioSegment)
			//}
		default:
			log.Println("未知消息类型:", messageType)
		}
	}
}

func reportStream(sr *schema.StreamReader[*schema.Message], conn *websocket.Conn) {
	defer sr.Close()

	var combinedMessages string
	if sr == nil {
		return
	}
	for {
		message, err := sr.Recv()

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("recv failed: %v", err)
			return
		}
		// 将消息拼接到 combinedMessages
		combinedMessages += message.Content
		// 将消息发送到 WebSocket 连接
		err = conn.WriteJSON(message)
		if err != nil {
			log.Printf("write to websocket failed: %v", err)
			return
		}
	}
	go context.SaveContext("ai", combinedMessages) //异步保存对话历史
	//调用tts
	audioData, err := tts.Execute(combinedMessages, "zh-CN-XiaoxiaoNeural")
	if err != nil {
		log.Println("TTS执行失败:", err)
	}
	//bufferSize := 1024
	//for i := 0; i < len(audioData); i += bufferSize {
	//	end := i + bufferSize
	//	if end > len(audioData) {
	//		end = len(audioData)
	//	}
	//	audioChunk := audioData[i:end]
	//	audioMessage := map[string]interface{}{
	//		"type":    "audio",
	//		"content": audioChunk,
	//	}
	//	err = conn.WriteJSON(audioMessage)
	//	if err != nil {
	//		log.Println("发送音频数据错误:", err)
	//		break
	//	}
	//}
	//endMessage := map[string]interface{}{
	//	"type":    "audio",
	//	"content": "",
	//}
	err = os.MkdirAll("cache", 0755) // Create the directory if it doesn't exist
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}
	filename := uuid.New().String()
	path := "cache/" + filename + ".mp3"
	writeMediaErr := os.WriteFile(path, audioData, 0644)
	if writeMediaErr != nil {
		fmt.Println(writeMediaErr)
	}
	mp3Data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("Failed to read MP3 file:", err)
		return
	}

	// 发送 MP3 文件数据流
	err = conn.WriteMessage(websocket.BinaryMessage, mp3Data)
	if err != nil {
		log.Println("Failed to send MP3 data:", err)
	}
	// 输出拼接后的消息到日志
	log.Printf("流式回复:\n%s", combinedMessages)
}
