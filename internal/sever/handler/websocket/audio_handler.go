package websocket

import (
	"github.com/gorilla/websocket"
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
	"go.uber.org/zap"
	"virtugo/internal/asr"
	"virtugo/internal/config"
	"virtugo/internal/kws"
	model "virtugo/internal/sever/message_model"
	"virtugo/logs"
)

type AudioHandler struct {
	sessionID  int
	printed    bool
	kwsSpotter *sherpa.KeywordSpotter
	asr        asr.Recognizer
	vad        *sherpa.VoiceActivityDetector
	room       *Room
	conn       *websocket.Conn
	username   string
}

func NewAudioHandler(username string, sessionID int, room *Room, conn *websocket.Conn) *AudioHandler {
	return &AudioHandler{
		sessionID:  sessionID,
		kwsSpotter: kws.NewKeyWordSpotter(),
		asr:        asr.NewASR("sherpa"),
		room:       room,
		conn:       conn,
		vad:        asr.NewVad(),
		printed:    false,
		username:   username,
	}
}

func (a *AudioHandler) AddAudioPiece(audioPiece []float32) {
	a.vad.AcceptWaveform(audioPiece)
	if a.vad.IsSpeech() {
		if !a.printed {
			a.printed = true
			interruptMsg := map[string]interface{}{
				"type":    "interrupt",
				"content": "interrupt",
			}
			logs.Logger.Debug("发送打断消息")
			a.conn.WriteJSON(interruptMsg)
		}
	} else {
		a.printed = false
	}
	for !a.vad.IsEmpty() {
		speechSegment := a.vad.Front()
		a.vad.Pop()

		audioWave := &sherpa.Wave{}
		audioWave.Samples = speechSegment.Samples
		audioWave.SampleRate = 16000

		if config.Cfg.KeyWordIsEnable {
			stream := sherpa.NewKeywordStream(a.kwsSpotter)
			defer sherpa.DeleteOnlineStream(stream)
			stream.AcceptWaveform(audioWave.SampleRate, audioWave.Samples)
			is_detected := false
			for a.kwsSpotter.IsReady(stream) {
				a.kwsSpotter.Decode(stream)
				result := a.kwsSpotter.GetResult(stream)
				if result.Keyword != "" {
					a.kwsSpotter.Reset(stream)
					logs.Logger.Debug("关键词检测到", zap.String("result", result.Keyword))
					is_detected = true
					break
				}
			}
			if !is_detected {
				logs.Logger.Debug("未检测到关键词")
				break
			}
		}

		str, err := a.asr.Recognize(audioWave)
		if err != nil {
			logs.Logger.Error("asr识别失败", zap.Error(err))
		}
		logs.Logger.Info("asr识别结果", zap.String("result", str))
		if str == "" || str == " " {
			logs.Logger.Warn("asr识别结果为空")
			break
		}
		input := map[string]any{
			"session_id": a.sessionID,
			"question":   str,
			"username":   a.username,
		}
		resp := model.WebsocketRespMessage{
			Type:     "user_audio_input",
			RoleName: a.username,
			Content:  str,
		}
		a.conn.WriteJSON(resp)

		go a.room.Speak(input)

	}
}
