package context

import (
	"context"
	"errors"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"virtugo/internal/config"
	"virtugo/internal/tts"
	"virtugo/logs"
)

type ServiceContext struct {
	chatModel     *openai.ChatModel //Todo后续会改进为多个模型
	currentCtx    context.Context
	currentCancel context.CancelFunc
	//Template  prompt.ChatTemplate
	Message  []*schema.Message
	Agent    *react.Agent
	Chain    *compose.Chain[string, *schema.Message]
	app      compose.Runnable[string, *schema.Message]
	conn     *websocket.Conn
	stopFlag atomic.Bool
	mu       sync.Mutex
}

func (s *ServiceContext) InitServiceContext(conn1 *websocket.Conn) {
	s.conn = conn1
	ctx := context.Background()

	s.chatModel, _ = openai.NewChatModel(ctx, &openai.ChatModelConfig{
		//Model:  "gpt-4o-mini", // 使用的模型版本
		//APIKey:  os.Getenv("OPENAI_API_KEY"),
		APIKey:  config.Cfg.ModelInfo.ApiKey,
		BaseURL: config.Cfg.ModelInfo.BaseUrl,
		Model:   config.Cfg.ModelInfo.ModelName,

		Temperature: &config.Cfg.ModelInfo.Temperature,
	})
	persona := config.Cfg.ModelInfo.Persona
	saveMemoryTool := GetSaveMemoryTool()
	//queryTextTool := GetQueryTextTool()
	changeLanguageTool := ChangeLanguage()
	SleepTool := Sleep()
	tools := []tool.BaseTool{}
	tools = append(tools, saveMemoryTool)
	tools = append(tools, changeLanguageTool)
	tools = append(tools, SleepTool)
	mcpTools := config.GetEinoTools()
	tools = append(tools, mcpTools...)
	toolsConfig := compose.ToolsNodeConfig{
		Tools: tools,
	}
	var err error

	s.Agent, err = react.NewAgent(ctx, &react.AgentConfig{
		Model:              s.chatModel,
		ToolsConfig:        toolsConfig,
		MessageModifier:    react.NewPersonaModifier(persona),
		ToolReturnDirectly: map[string]struct{}{"go_to_sleep": {}},
		StreamToolCallChecker: func(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (bool, error) {
			defer sr.Close()
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
						err = conn1.WriteMessage(websocket.BinaryMessage, mp3Data)

						if err != nil {
							logs.Logger.Error("发送MP3文件失败", zap.Error(err))
						}
					}
				}
			}
			lan := config.Cfg.Language
			for {
				message, err := sr.Recv()

				if errors.Is(err, io.EOF) {
					if len(buffer) > 1 {
						logs.Logger.Debug("LLM最后回复", zap.String("content", buffer))
						if lan == "zh" {
							tts(buffer)
						}
						if len(combinedMessages) > 1 {
							go SaveContext("ai", combinedMessages)
						}
					}
					return false, nil
				}
				if err != nil {
					logs.Logger.Error("接收消息流片段失败", zap.Error(err))
				}
				if len(message.ToolCalls) > 0 {
					logs.Logger.Debug("ToolCalls", zap.Any("ToolCalls", message.ToolCalls))
					if len(combinedMessages) > 1 {
						go SaveContext("ai", combinedMessages)
					}
					return true, nil
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
					err = conn1.WriteJSON(message)
					if err != nil {
						logs.Logger.Error("发送消息到 WebSocket 失败", zap.Error(err))
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
					err = conn1.WriteJSON(message)
					if err != nil {
						logs.Logger.Error("发送消息到 WebSocket 失败", zap.Error(err))
					}
				}
			}
		},
	})
	if err != nil {
		log.Println("创建Agent失败:", err)
	}
	//modelHandler := &callbacks.ModelCallbackHandler{
	//	OnStart: nil,
	//	OnEnd:   nil,
	//	OnEndWithStreamOutput: func(ctx context.Context, runInfo *callbacks2.RunInfo, output *schema.StreamReader[*model.CallbackOutput]) context.Context {
	//
	//		return ctx
	//	},
	//	OnError: nil,
	//}
	//toolHandler := &callbacks.ToolCallbackHandler{
	//	OnStart:               nil,
	//	OnEnd:                 nil,
	//	OnEndWithStreamOutput: nil,
	//	OnError:               nil,
	//}
	//Handler := react.BuildAgentCallback(modelHandler, toolHandler)

	agentLambda, _ := compose.AnyLambda(s.Agent.Generate, s.Agent.Stream, nil, nil)
	loadMomeryLamda := InitLoadMemory()

	SetLanguage := SetLanguage()
	chatTemplate := NewChatTemplate()
	s.Chain = compose.NewChain[string, *schema.Message]()
	s.Chain.
		AppendLambda(loadMomeryLamda).
		AppendLambda(SetLanguage).
		AppendChatTemplate(chatTemplate).
		AppendLambda(agentLambda)

	s.app, err = s.Chain.Compile(context.Background())
	if err != nil {
		log.Println("编译Chain失败:", err)
	}

}

func (s *ServiceContext) Generate() (*schema.Message, error) {

	return s.chatModel.Generate(s.currentCtx, s.Message)

}
func (s *ServiceContext) Stream(input string) (*schema.StreamReader[*schema.Message], error) {
	s.mu.Lock()
	// 如果存在旧的context，先取消它
	if s.currentCancel != nil {
		s.currentCancel()
	}
	s.currentCtx, s.currentCancel = context.WithCancel(context.Background())
	ctx := s.currentCtx
	s.mu.Unlock()

	return s.app.Stream(ctx, input)
}
func (s *ServiceContext) StopStream() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentCancel != nil {
		s.currentCancel()
		s.currentCancel = nil
	}
}
