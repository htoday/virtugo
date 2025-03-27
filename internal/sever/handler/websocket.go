package handler

import (
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"virtugo/internal/asr"
	"virtugo/internal/sever/context"
	model "virtugo/internal/sever/message_model"
	"virtugo/internal/tts"
	"virtugo/logs"
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
		logs.Logger.Error("WebSocket 升级失败", zap.Error(err))
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
	logs.Logger.Info("创建连接成功")
	for {
		// 读取消息
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			logs.Logger.Error("读取消息错误", zap.Error(err))
			break
		}

		switch messageType {
		case websocket.TextMessage: // 文本消息
			logs.Logger.Info("收到文本消息" + string(msg))
			var inputMessage model.TextMessage
			err := json.Unmarshal(msg, &inputMessage)
			if err != nil {
				logs.Logger.Warn("解析消息错误", zap.Error(err))
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
					logs.Logger.Error("获取LLM流式回复失败", zap.Error(err))
					continue
				}
				reportStream(streamResult, conn)
				mutex.Unlock()
			case "image":
				fmt.Println("处理图片消息:", inputMessage.Content)
			default:
				logs.Logger.Warn("未知文本消息类型", zap.String("type", inputMessage.Type))
			}
			//conn.WriteMessage(websocket.TextMessage, []byte(response))

		case websocket.BinaryMessage:
			// 识别音频
			// 解析 PCM 数据
			pcmData := samplesInt16ToFloat(msg)

			vad.AcceptWaveform(pcmData)

			if vad.IsSpeech() && !printed {
				printed = true
				logs.Logger.Debug("检测到语音")
				interrupt_msg := map[string]interface{}{
					"type":    "interrupt",
					"content": "interrupt",
				}
				logs.Logger.Debug("发送打断消息")
				err = conn.WriteJSON(interrupt_msg)
				if err != nil {
					logs.Logger.Warn("发送中断消息失败", zap.Error(err))
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
						logs.Logger.Error("获取LLM流式回复失败", zap.Error(err))
					}
					reportStream(streamResult, conn)
					mutex.Unlock()
				}()

				k += 1
			}
		default:
			logs.Logger.Warn("未知消息类型", zap.Int("type", messageType))
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
			logs.Logger.Error("接收消息流片段失败", zap.Error(err))
			return
		}
		// 将消息拼接到 combinedMessages
		combinedMessages += message.Content
		// 将消息发送到 WebSocket 连接
		err = conn.WriteJSON(message)
		if err != nil {
			logs.Logger.Error("发送消息到 WebSocket 失败", zap.Error(err))
			return
		}
	}
	go context.SaveContext("ai", combinedMessages) //异步保存对话历史
	//调用tts

	tts_service, err := tts.NewTTS()
	if err != nil {
		logs.Logger.Error("TTS初始化失败", zap.Error(err))
	} else {
		path, err := tts_service.Generate(combinedMessages)
		if err != nil {
			logs.Logger.Error("TTS执行失败", zap.Error(err))
		} else {
			mp3Data, err := ioutil.ReadFile(path)
			if err != nil {
				logs.Logger.Error("读取MP3文件失败", zap.Error(err))
			}

			// 发送 MP3 文件数据流
			err = conn.WriteMessage(websocket.BinaryMessage, mp3Data)

			if err != nil {
				logs.Logger.Error("发送MP3文件失败", zap.Error(err))
			}
		}
	}
	//audioData, err := tts.Execute(combinedMessages, "zh-CN-XiaoxiaoNeural")
	//if err != nil {
	//	logs.Logger.Error("TTS执行失败", zap.Error(err))
	//}else{
	//
	//}
	// 保存 MP3 文件
	//err = os.MkdirAll("cache", 0755) // Create the directory if it doesn't exist
	//if err != nil {
	//	logs.Logger.Error("创建cache目录失败", zap.Error(err))
	//	return
	//}
	//
	//
	//filename := uuid.New().String()
	//path := "cache/" + filename + ".mp3"
	//writeMediaErr := os.WriteFile(path, audioData, 0644)
	//if writeMediaErr != nil {
	//	logs.Logger.Error("写入MP3文件失败", zap.Error(writeMediaErr))
	//}

	// 输出拼接后的消息到日志
	logs.Logger.Info("LLM回复", zap.String("content", combinedMessages))
}
