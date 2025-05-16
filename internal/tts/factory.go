package tts

import (
	"errors"
	"strings"
	"virtugo/internal/config"
)

type TTS interface {
	Generate(text string) (string, error)
}

func NewTTS(roleName string) (TTS, error) {
	serviceType := config.Cfg.Models[roleName].TTS.ServiceType

	switch {
	case strings.Contains(serviceType, "edge"):
		//logs.Logger.Debug("使用 EdgeTTS")
		edge_tts_voice := config.Cfg.Models[roleName].TTS.EdgeVoice
		if edge_tts_voice == "" {
			edge_tts_voice = "zh-CN-XiaoxiaoNeural"
		}
		return NewEdgeTTS(edge_tts_voice), nil
	case strings.Contains(serviceType, "fish"):
		//logs.Logger.Debug("使用 FishAudioTTS")
		fish_audio_voice := config.Cfg.Models[roleName].TTS.FishAudioVoice
		fish_audio_key := config.Cfg.Models[roleName].TTS.FishAudioKey
		if fish_audio_key == "" {
			return nil, errors.New("缺少配置项 fish_audio_key")
		}
		if fish_audio_voice == "" {
			return nil, errors.New("缺少配置项 fish_audio_voice")
		}
		return NewFishAudioTTS(fish_audio_voice, fish_audio_key), nil
	default:
		edge_tts_voice := config.Cfg.Models[roleName].TTS.EdgeVoice
		if edge_tts_voice == "" {
			edge_tts_voice = "zh-CN-XiaoxiaoNeural"
		}
		return NewEdgeTTS(edge_tts_voice), nil
	}
	return nil, errors.New("未知的 TTS 服务类型")
}
