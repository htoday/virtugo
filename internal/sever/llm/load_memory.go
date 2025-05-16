package llm

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
	"time"
	"virtugo/internal/config"
	"virtugo/internal/dao"
	"virtugo/logs"
)

func InitLoadMemory() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		// 从input中获取会话ID
		sessionID, ok := input["session_id"].(int)
		username := input["username"].(string)
		//logs.Logger.Debug("保存用户消息:", zap.String("username", username))
		if !ok {
			// 如果没有提供有效的会话ID，返回错误或使用默认处理
			logs.Logger.Error("没有提供有效的会话ID")
			sessionID = 1 // 使用默认会话ID
		}

		// 从input中获取用户输入的文本
		inputText, ok := input["question"].(string)
		if !ok {
			inputText = ""
			logs.Logger.Error("没有提供有效的输入文本")
		}

		maxLength := config.Cfg.History.MaxLength
		if maxLength <= 0 {
			maxLength = 10 // 设置默认值
		}
		// 获取特定会话的最近n条消息
		messages, err := dao.GetSessionMessages(dao.SqliteDB, sessionID, maxLength)
		if err != nil {
			logs.Logger.Error("获取会话历史记录失败", zap.Error(err), zap.Int("session_id", sessionID))
			return input, err
		}
		// 创建一个切片来存储消息历史
		var chatHistory []*schema.Message

		for i := 0; i <= len(messages)-1; i++ {
			msg := messages[i]
			role, ok := msg["role"].(string)
			if !ok {
				logs.Logger.Error("无效的消息角色", zap.Any("message", msg))
				continue
			}

			var timeMsg string
			switch v := msg["created_at"].(type) {
			case string:
				timeMsg = v
			case time.Time:
				timeMsg = v.Format(time.RFC3339)
			default:
				logs.Logger.Error("无效的消息创建时间", zap.Any("message", msg))
				continue
			}
			content, ok := msg["content"].(string)
			if !ok {
				logs.Logger.Error("无效的消息内容", zap.Any("message", msg))
				continue
			}
			timeMsg = timeMsg[:19] // 截取到秒级别

			senderName, ok := msg["senderName"].(string)
			if !ok {
				logs.Logger.Warn("无效的消息发送者名称", zap.Any("message", msg))
			}
			jsonContent := HistoryMessage{
				SenderName: senderName,
				Content:    content,
				Time:       timeMsg,
			}
			bytesJsonContent, err := json.Marshal(jsonContent)
			if err != nil {
				logs.Logger.Error("json序列化失败", zap.Error(err))
				continue
			}
			strJsonContent := string(bytesJsonContent)
			// 根据角色创建相应的消息实例
			var schemaMsg *schema.Message
			if role == "user" {
				schemaMsg = schema.UserMessage(strJsonContent)
			} else if role == "assistant" || role == "ai" {
				schemaMsg = schema.AssistantMessage(strJsonContent, nil)
			} else if role == "system" {
				schemaMsg = schema.SystemMessage(strJsonContent)
			} else {
				logs.Logger.Error("未知的角色", zap.String("role", role))
				continue
			}

			//schemaMsg.Extra["sender"] = senderName
			// 将消息添加到 chatHistory 切片
			chatHistory = append(chatHistory, schemaMsg)
		}

		// 将当前用户输入保存到数据库
		if inputText != "" {
			// 保存用户当前的输入到数据库
			go func() {
				_, err := SaveMessageToSession(dao.SqliteDB, sessionID, "user", username, inputText, 0)
				if err != nil {
					logs.Logger.Error("保存用户消息失败", zap.Error(err))
				}
			}()
		}

		input["chat_history"] = chatHistory
		input["question"] = inputText
		return input, nil
	})
}

func InitGroupLoadMemory() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		// 从input中获取会话ID
		sessionID, ok := input["session_id"].(int)
		if !ok {
			// 如果没有提供有效的会话ID，返回错误或使用默认处理
			logs.Logger.Error("没有提供有效的会话ID")
			sessionID = 1 // 使用默认会话ID
		}
		// 获取特定会话的最近10条消息
		logs.Logger.Info("正在获取会话历史记录", zap.Int("session_id", sessionID))
		messages, err := dao.GetSessionMessages(dao.SqliteDB, sessionID, 10)
		if err != nil {
			logs.Logger.Error("获取会话历史记录失败", zap.Error(err), zap.Int("session_id", sessionID))
			return input, err
		}
		// 创建一个切片来存储消息历史
		var chatHistory []*schema.Message

		for i := 0; i <= len(messages)-1; i++ {
			msg := messages[i]
			role, ok := msg["role"].(string)
			if !ok {
				logs.Logger.Error("无效的消息角色", zap.Any("message", msg))
				continue
			}
			var timeMsg string
			switch v := msg["created_at"].(type) {
			case string:
				timeMsg = v
			case time.Time:
				timeMsg = v.Format(time.RFC3339)
			default:
				logs.Logger.Error("无效的消息创建时间", zap.Any("message", msg))
				continue
			}
			timeMsg = timeMsg[:19] // 截取到秒级别暂时不用
			content, ok := msg["content"].(string)
			if !ok {
				logs.Logger.Error("无效的消息内容", zap.Any("message", msg))
				continue
			}
			senderName, ok := msg["senderName"].(string)
			if !ok {
				logs.Logger.Warn("无效的消息发送者名称", zap.Any("message", msg))
			}
			jsonContent := HistoryMessage{
				SenderName: senderName,
				Content:    content,
				Time:       timeMsg,
			}
			bytesJsonContent, err := json.Marshal(jsonContent)
			if err != nil {
				logs.Logger.Error("json序列化失败", zap.Error(err))
				continue
			}
			strJsonContent := string(bytesJsonContent)
			// 根据角色创建相应的消息实例
			var schemaMsg *schema.Message
			if role == "user" {
				schemaMsg = schema.UserMessage(strJsonContent)
			} else if role == "assistant" || role == "ai" {
				schemaMsg = schema.AssistantMessage(strJsonContent, nil)
			} else if role == "system" {
				schemaMsg = schema.SystemMessage(strJsonContent)
			} else {
				logs.Logger.Error("未知的角色", zap.String("role", role))
				continue
			}
			chatHistory = append(chatHistory, schemaMsg)
		}
		input["chat_history"] = chatHistory
		return input, nil
	})
}
