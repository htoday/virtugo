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
	"strings"
	"sync"
	"time"
	"virtugo/internal/asr"
	"virtugo/internal/config"
	"virtugo/internal/kws"
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

	//初始化服务上下文
	serviceContext := context.ServiceContext{}
	serviceContext.InitServiceContext()

	//初始化重置定时器
	rt := NewResettableTimer(30*time.Second, func() {
		serviceContext.WorkStage = "sleep"
	})

	//初始化语音识别
	recAsr := asr.InitOfflineASR()

	//初始化VAD
	vadConfig := asr.InitVad()
	var bufferSizeInSeconds float32 = 20
	vad := sherpa.NewVoiceActivityDetector(&vadConfig, bufferSizeInSeconds)

	kwsConfig := kws.InitKeyWordSpot()
	spotter := sherpa.NewKeywordSpotter(&kwsConfig)

	var mutex sync.Mutex

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

				streamResult, err := serviceContext.Stream(inputMessage.Content)
				if err != nil {
					logs.Logger.Error("获取LLM流式回复失败", zap.Error(err))
					continue
				}
				reportStream(streamResult, conn)

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

				if serviceContext.WorkStage == "sleeping" {
					stream := sherpa.NewKeywordStream(spotter)
					defer sherpa.DeleteOnlineStream(stream)
					stream.AcceptWaveform(audio.SampleRate, audio.Samples)
					isDetected := false
					for spotter.IsReady(stream) {
						spotter.Decode(stream)
						result := spotter.GetResult(stream)
						if result.Keyword != "" {
							spotter.Reset(stream)
							logs.Logger.Info("检测到关键词", zap.String("keyword", result.Keyword))
							isDetected = true
							go rt.Start()
							serviceContext.WorkStage = "work"
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
						}
					}
					if !isDetected {
						logs.Logger.Debug("未检测到关键词")
					}
				} else {
					go rt.Reset()
					go func() {
						text := recognizeSpeech(recAsr, audio)

						//mutex.Lock()
						streamResult, err := serviceContext.Stream(text)
						if err != nil {
							logs.Logger.Error("获取LLM流式回复失败", zap.Error(err))
						}
						reportStream(streamResult, conn)
						//mutex.Unlock()
					}()
				}

				k += 1
			}
		default:
			logs.Logger.Warn("未知消息类型", zap.Int("type", messageType))
		}
	}
}

func reportStream(sr *schema.StreamReader[*schema.Message], conn *websocket.Conn) {

	var combinedMessages string
	var buffer string
	var isChinese bool
	isChinese = false
	buffer = ""
	tts := func(text string) {
		tts_service, err := tts.NewTTS()
		if err != nil {
			logs.Logger.Error("TTS初始化失败", zap.Error(err))
		} else {
			path, err := tts_service.Generate(text)
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
	}

	if sr == nil {

		return
	}
	lan := config.Cfg.Language
	defer sr.Close()
	for {
		message, err := sr.Recv()

		if err == io.EOF {
			if len(buffer) > 1 {
				logs.Logger.Debug("LLM最后回复", zap.String("content", buffer))
				if lan == "zh" {
					tts(buffer)
				}
			}
			break
		}
		if err != nil {
			logs.Logger.Error("接收消息流片段失败", zap.Error(err))
			return
		}

		// 将消息拼接到 combinedMessages
		combinedMessages += message.Content
		buffer += message.Content

		if lan == "zh" {
			if strings.Contains(buffer, "|") {
				parts := strings.Split(buffer, "|")
				// 处理完整的部分
				for i := 0; i < len(parts)-1; i++ {
					logs.Logger.Debug("LLM回复", zap.String("content", parts[i]))
					tts(parts[i])
				}
				// 剩下最后一个部分是可能未完成的内容，留着
				buffer = parts[len(parts)-1]

			}
			if strings.Contains(buffer, "｜") {
				parts := strings.Split(buffer, "｜")
				// 处理完整的部分
				for i := 0; i < len(parts)-1; i++ {
					logs.Logger.Debug("LLM回复", zap.String("content", parts[i]))
					tts(parts[i])
				}
				// 剩下最后一个部分是可能未完成的内容，留着
				buffer = parts[len(parts)-1]
			}
			// 将消息发送到 WebSocket 连接
			err = conn.WriteJSON(message)
			if err != nil {
				logs.Logger.Error("发送消息到 WebSocket 失败", zap.Error(err))
				return
			}
			continue
		}
		if !isChinese {
			if strings.Contains(buffer, "|") {
				parts := strings.Split(buffer, "|")
				// 处理完整的部分
				for i := 0; i < len(parts)-1; i++ {
					logs.Logger.Debug("LLM回复", zap.String("content", parts[i]))
					tts(parts[i])
				}
				// 剩下最后一个部分是可能未完成的内容，留着
				buffer = parts[len(parts)-1]

			}
			if strings.Contains(buffer, "｜") {
				parts := strings.Split(buffer, "｜")
				// 处理完整的部分
				for i := 0; i < len(parts)-1; i++ {
					logs.Logger.Debug("LLM回复", zap.String("content", parts[i]))
					tts(parts[i])
				}
				// 剩下最后一个部分是可能未完成的内容，留着
				buffer = parts[len(parts)-1]
			}
			if strings.Contains(buffer, "——") {
				parts := strings.Split(buffer, "——")
				// 处理完整的部分
				for i := 0; i < len(parts)-1; i++ {
					logs.Logger.Debug("LLM回复", zap.String("content", parts[i]))
					tts(parts[i])
				}
				// 剩下最后一个部分是可能未完成的内容，留着
				buffer = parts[len(parts)-1]
				isChinese = true
			}
		} else {
			// 将消息发送到 WebSocket 连接
			err = conn.WriteJSON(message)
			if err != nil {
				logs.Logger.Error("发送消息到 WebSocket 失败", zap.Error(err))
				return
			}
		}
	}

	go context.SaveContext("ai", combinedMessages) //异步保存对话历史
	//调用tts

	// 输出拼接后的消息到日志
	logs.Logger.Info("LLM回复", zap.String("content", combinedMessages))
}
