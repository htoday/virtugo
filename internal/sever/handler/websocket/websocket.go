package websocket

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"virtugo/internal/config"
	"virtugo/internal/sever/global"
	"virtugo/internal/sever/llm"
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

	ttsQueue := make(chan llm.TTSRequest, 100)
	defer close(ttsQueue)
	//在 websocket 内部消费队列，串行执行 TTS 并发送
	ttsFunc := func(text string, msgID int64, index int32, sender string) {
		tts_service, err := tts.NewTTS(sender)
		if err != nil {
			logs.Logger.Error("TTS初始化失败", zap.Error(err))
		} else {
			if len(text) < 1 {
				logs.Logger.Error("文本为空，不执行TTS")
				return
			}
			if text == "EOF" {
				resp := llm.TTSMessage{
					Index:     index,
					MessageID: msgID,
					Text:      text,
				}
				err = conn.WriteJSON(resp)
				if err != nil {
					logs.Logger.Error("发送EOF文本数据到 WebSocket 失败", zap.Error(err))
				}
				return
			}
			path, err := tts_service.Generate(text)
			if err != nil {
				logs.Logger.Error("TTS执行失败", zap.Error(err))
			} else {
				logs.Logger.Info("tts成功", zap.String("文本：", text))
				mp3Data, err := ioutil.ReadFile(path)
				if err != nil {
					logs.Logger.Error("读取MP3文件失败", zap.Error(err))
				}

				resp := llm.TTSMessage{
					Index:      index,
					MessageID:  msgID,
					Text:       text,
					Audio:      mp3Data,
					SenderName: sender,
				}
				err = conn.WriteJSON(resp)
				if err != nil {
					logs.Logger.Error("发送音频和文本数据到 WebSocket 失败", zap.Error(err))
				}
			}
		}
	}

	go func() {
		for req := range ttsQueue {
			ttsFunc(req.Text, req.MessageID, req.Index, req.Sender)

		}
	}()
	//初始化llm
	room := NewRoom(ttsQueue)
	Rooms.Add(room.RoomID, room)
	defer Rooms.Delete(room.RoomID)

	global.WorkState = "sleeping"

	var audioHandler *AudioHandler

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
			var inputMessage model.WebsocketReqMessage
			err := json.Unmarshal(msg, &inputMessage)
			if err != nil {
				logs.Logger.Warn("解析消息错误", zap.Error(err))
			}
			switch inputMessage.Type {

			case "join":
				logs.Logger.Debug("加入房间", zap.String("", inputMessage.RoleName))
				if inputMessage.RoleName == "" {
					logs.Logger.Error("角色名称不能为空")
					resp := model.WebsocketRespMessage{
						Type:     "error",
						RoleName: inputMessage.RoleName,
						Content:  "角色名称不能为空",
					}
					conn.WriteJSON(resp)
					continue
				}
				if config.Cfg.Models[inputMessage.RoleName].ModelInfo.APIKey == "" {
					logs.Logger.Error("角色名称不存在或信息不完整")
					resp := model.WebsocketRespMessage{
						Type:     "error",
						RoleName: inputMessage.RoleName,
						Content:  "角色名称不存在或信息不完整",
					}
					conn.WriteJSON(resp)
					continue
				}
				room.AddAI(inputMessage.RoleName)
				resp := model.WebsocketRespMessage{
					Type:     "success",
					RoleName: inputMessage.RoleName,
					Content:  "inputMessage.RoleName 加入房间成功",
				}
				conn.WriteJSON(resp)
			case "exit":
				logs.Logger.Debug("退出房间", zap.String("", inputMessage.RoleName))
				if inputMessage.RoleName == "" {
					logs.Logger.Error("角色名称不能为空")
					resp := model.WebsocketRespMessage{
						Type:     "error",
						RoleName: inputMessage.RoleName,
						Content:  "角色名称不能为空",
					}
					conn.WriteJSON(resp)
					continue
				}
				if config.Cfg.Models[inputMessage.RoleName].ModelInfo.APIKey == "" {
					logs.Logger.Error("角色名称不存在或信息不完整")
					resp := model.WebsocketRespMessage{
						Type:     "error",
						RoleName: inputMessage.RoleName,
						Content:  "角色名称不存在或信息不完整",
					}
					conn.WriteJSON(resp)
					continue
				}
				room.ExitAI(inputMessage.RoleName)
				resp := model.WebsocketRespMessage{
					Type:     "success",
					RoleName: inputMessage.RoleName,
					Content:  "inputMessage.RoleName 退出房间成功",
				}
				conn.WriteJSON(resp)
			case "play_done":
				logs.Logger.Debug("收到播放完成消息", zap.String("msgID", inputMessage.Content))
				playDoneMsgIDStr := inputMessage.Content
				playDoneMsgID, err := strconv.ParseInt(playDoneMsgIDStr, 10, 64)
				if err != nil {
					logs.Logger.Error("解析play_done消息错误", zap.Error(err))
					continue
				}
				if room != nil {
					room.PlayDone <- playDoneMsgID
				}
			case "start_call":
				if inputMessage.SessionID == 0 {
					logs.Logger.Error("session_id不能为空")
					continue
				}
				username := c.Value("username").(string)
				audioHandler = NewAudioHandler(username, inputMessage.SessionID, room, conn)
			case "stop_call":

			case "interrupt":
				logs.Logger.Info("收到打断消息")
				room.StopTalk()

			case "text":
				logs.Logger.Debug("收到文本消息")
				chainInput := map[string]any{
					"session_id": inputMessage.SessionID,
					"username":   c.Value("username"),
					"question":   inputMessage.Content,
				}
				logs.Logger.Debug("session_id", zap.Int("session_id", inputMessage.SessionID))
				interruptMsg := map[string]interface{}{
					"type":    "interrupt",
					"content": "interrupt",
				}
				logs.Logger.Debug("发送打断消息")
				conn.WriteJSON(interruptMsg)
				go func() {
					room.Speak(chainInput)
				}()
			case "nospeek":

			case "image":
				fmt.Println("处理图片消息:", inputMessage.Content)
			default:
				logs.Logger.Warn("未知文本消息类型", zap.String("type", inputMessage.Type))
			}
			//conn.WriteMessage(websocket.WebsocketReqMessage, []byte(response))

		case websocket.BinaryMessage:
			// 识别音频
			// 解析 PCM 数据
			pcmData := samplesInt16ToFloat(msg)
			if audioHandler == nil {
				logs.Logger.Error("音频处理器未初始化")
				continue
			}
			audioHandler.AddAudioPiece(pcmData)
			continue
		}
	}
}
