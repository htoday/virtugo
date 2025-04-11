package context

import (
	"context"
	"github.com/cloudwego/eino/compose"
	"time"
	"virtugo/internal/config"
	"virtugo/logs"
)

func SetLanguage() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		var language string
		if config.Cfg.Language == "jp" {
			language = "请严格遵守格式，你的回答需要先用日语回答，再用相同意思的中文回答，两次回答中间用——分隔，例如 こんにちは——你好，如果有多句话，用｜分割，例如 一言目｜二言目 —— 第一句话｜第二句话"
			logs.Logger.Debug("日语回复")
		} else if config.Cfg.Language == "en" {
			language = "请严格遵守格式,你的回答需要先用英语回答，再用相同意思的中文回答，两次回答中间用——分隔,，例如 hello——你好，如果有多句话，用｜分割，例如 first sentence｜second sentence —— 第一句话｜第二句话"
			logs.Logger.Debug("英语回复")
		} else if config.Cfg.Language == "zh" {
			language = "你直接用中文回答，不需要翻译，例如 第一句话｜第二句话，不需要加——和另一种语言"
			logs.Logger.Debug("中文回复")
		} else {
			language = "请严格遵守格式，你需要用把回答翻译成不同的语言,你的回答需要先用日语，再用中文回答，中间用——分隔,而不是｜，例如 こんにちは——你好"
		}
		currentTime := time.Now()
		input["current_time"] = currentTime.Format("2006-01-02 15:04:05")
		input["language_config"] = language
		return input, nil
	})
}
