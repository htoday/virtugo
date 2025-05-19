package llm

import (
	"context"
	"github.com/cloudwego/eino/compose"
	"time"
	"virtugo/internal/config"
	"virtugo/logs"
)

func LoadPrompt() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {

		currentTime := time.Now()
		aiName, ok := input["ai_name"].(string)
		if !ok {
			logs.Logger.Error("没有找到有效的角色提示词")
		}
		input["current_time"] = currentTime.Format("2006-01-02 15:04:05")
		input["prompt"] = config.Cfg.Models[aiName].ModelInfo.SystemPrompt
		if input["prompt"] == "" {
			logs.Logger.Warn("没有找到有效的角色提示词")
			input["prompt"] = " "
		}
		return input, nil
	})
}
