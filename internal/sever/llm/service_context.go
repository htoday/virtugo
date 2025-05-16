package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/ddgsearch"
	"github.com/cloudwego/eino-ext/components/tool/wikipedia"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	callbacks2 "github.com/cloudwego/eino/utils/callbacks"
	"github.com/gorilla/websocket"
	"github.com/ollama/ollama/api"
	"go.uber.org/zap"
	"io"
	"log"
	"sync"
	"sync/atomic"
	"time"
	"virtugo/internal/config"
	"virtugo/internal/dao"
	"virtugo/logs"
)

type ChatChain struct {
	//chatModel     *openai.ChatModel //Todo后续会改进为多个模型
	currentCtx    context.Context
	currentCancel context.CancelFunc
	//Template  prompt.ChatTemplate
	Message []*schema.Message
	//Agent     *react.Agent
	//Chain     *compose.Chain[map[string]any, *schema.Message]
	app  compose.Runnable[map[string]any, *schema.Message]
	conn *websocket.Conn
	//stopFlag atomic.Bool
	mu       sync.Mutex
	ttsQueue chan TTSRequest
	roleName string
}

func (s *ChatChain) InitChain(queue chan TTSRequest, modelName string) {
	s.ttsQueue = queue
	s.roleName = config.Cfg.Models[modelName].ModelInfo.RoleName
	var err error
	ctx := context.Background()
	temperature := config.Cfg.Models[modelName].ModelInfo.Temperature
	apiType := config.Cfg.Models[modelName].ModelInfo.ApiType
	var chatModel model.ToolCallingChatModel
	switch apiType {
	case "openai":
		chatModel, err = openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey:      config.Cfg.Models[modelName].ModelInfo.APIKey,
			BaseURL:     config.Cfg.Models[modelName].ModelInfo.BaseURL,
			Model:       config.Cfg.Models[modelName].ModelInfo.ModelName,
			Temperature: &temperature,
		})
		if err != nil {
			logs.Logger.Fatal("创建OpenAI模型失败", zap.Error(err))
		}
	case "ollama":
		chatModel, err = ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
			BaseURL: config.Cfg.Models[modelName].ModelInfo.BaseURL,
			Timeout: 30 * time.Second,
			Model:   config.Cfg.Models[modelName].ModelInfo.ModelName,
			Options: &api.Options{
				Temperature: temperature,
			},
		})
		if err != nil {
			logs.Logger.Fatal("创建Ollama模型失败", zap.Error(err))
		}
	default:
		logs.Logger.Error("不支持的模型类型,默认使用openai", zap.String("api_type", apiType))
		chatModel, err = openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey:      config.Cfg.Models[modelName].ModelInfo.APIKey,
			BaseURL:     config.Cfg.Models[modelName].ModelInfo.BaseURL,
			Model:       config.Cfg.Models[modelName].ModelInfo.ModelName,
			Temperature: &temperature,
		})
		if err != nil {
			logs.Logger.Fatal("创建OpenAI模型失败", zap.Error(err))
		}

	}

	fmt.Println("当前模型:", config.Cfg.Models[modelName].ModelInfo.ModelName)
	fmt.Println("当前模型APIKey:", config.Cfg.Models[modelName].ModelInfo.APIKey)
	persona := config.Cfg.Models[modelName].ModelInfo.Persona

	//SleepTool := Sleep()
	tools := []tool.BaseTool{}
	//tools = append(tools, SleepTool)

	//cfg := &ddgsearch.Config{
	//	Headers: map[string]string{
	//		"User-Agent": "MyApp/ 1.0",
	//	},
	//	Timeout:    10 * time.Second,
	//	Cache:      true,
	//	MaxRetries: 3,
	//}
	//DDG, err := ddgsearch.New(cfg)
	//if err != nil {
	//	logs.Logger.Error("创建DDG失败", zap.Error(err))
	//}
	//but, err := browseruse.NewBrowserUseTool(ctx, &browseruse.Config{
	//	Headless:           false,
	//	DisableSecurity:    false,
	//	ExtraChromiumArgs:  nil,
	//	ChromeInstancePath: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
	//	ProxyServer:        "",
	//	DDGSearchTool:      DDG,
	//	ExtractChatModel:   s.chatModel,
	//	Logf:               nil,
	//})
	//tools = append(tools, but)
	// Create configuration

	if config.Cfg.Models[modelName].Tools.DuckDuckGo.IsEnable {
		duckduckConfig := &duckduckgo.Config{
			MaxResults: 10, // Limit to return 3 results
			Region:     ddgsearch.RegionCN,
			DDGConfig: &ddgsearch.Config{
				Timeout:    5 * time.Second,
				Cache:      true,
				MaxRetries: 3,
			},
			TimeRange: ddgsearch.TimeRangeAll,
		}

		// Create search client
		duckduckTool, err := duckduckgo.NewTool(ctx, duckduckConfig)

		if err != nil {
			logs.Logger.Error("创建DuckDuckGo工具失败", zap.Error(err))
		} else {
			tools = append(tools, duckduckTool)
		}
	}
	if config.Cfg.Models[modelName].Tools.Wikipedia.IsEnable {
		userAengt := config.Cfg.Models[modelName].Tools.Wikipedia.UserAgent
		if userAengt != "" {
			config := &wikipedia.Config{
				UserAgent:   userAengt,
				DocMaxChars: 2000,
				Timeout:     5 * time.Second,
				TopK:        3,
				MaxRedirect: 3,
				Language:    "en",
			}
			// 创建搜索工具
			t, err := wikipedia.NewTool(ctx, config)
			if err != nil {
				logs.Logger.Error("创建Wikipedia工具失败", zap.Error(err))
			} else {
				tools = append(tools, t)
			}
		} else {
			config := &wikipedia.Config{
				UserAgent:   userAengt,
				DocMaxChars: 2000,
				Timeout:     15 * time.Second,
				TopK:        3,
				MaxRedirect: 3,
				Language:    "en",
			}
			// 创建搜索工具
			t, err := wikipedia.NewTool(ctx, config)
			if err != nil {
				logs.Logger.Error("创建Wikipedia工具失败", zap.Error(err))
			} else {
				tools = append(tools, t)
			}
		}
	}
	if config.Cfg.Models[modelName].Tools.MCPTool.IsEnable {
		mcpTools := config.GetEinoTools()
		tools = append(tools, mcpTools...)
	}

	toolsConfig := compose.ToolsNodeConfig{
		Tools: tools,
	}

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel:   chatModel,
		ToolsConfig:        toolsConfig,
		MessageModifier:    react.NewPersonaModifier(persona),
		ToolReturnDirectly: map[string]struct{}{"go_to_sleep": {}},
		StreamToolCallChecker: func(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (bool, error) {
			defer sr.Close()
			for {
				msg, err := sr.Recv()
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					return false, err
				}
				if len(msg.ToolCalls) > 0 {
					logs.Logger.Debug("检测到ToolCalls", zap.Any("ToolCalls", msg.ToolCalls))
					return true, nil
				}
			}
			logs.Logger.Debug("没有工具调用跳出循环")
			return false, nil
		},
	})
	if err != nil {
		log.Println("创建Agent失败:", err)
	}

	agentLambda, _ := compose.AnyLambda(agent.Generate, agent.Stream, nil, nil)
	loadMomeryLamda := InitLoadMemory()
	loadGroupMomeryLamda := InitGroupLoadMemory()
	SetLanguage := LoadPrompt()
	chatTemplate := NewChatTemplate()
	groupChatTemplate := NewGroupChatTemplate()
	branch1 := func(ctx context.Context, input map[string]any) (string, error) {
		if input["chat_type"] == "single" {
			return "single1", nil
		} else {
			return "group1", nil
		}
	}
	branch2 := func(ctx context.Context, input map[string]any) (string, error) {
		if input["chat_type"] == "single" {
			return "single2", nil
		} else {
			return "group2", nil
		}
	}
	Chain := compose.NewChain[map[string]any, *schema.Message]()
	Chain.
		AppendBranch(compose.NewChainBranch(branch1).AddLambda("single1", loadMomeryLamda).AddLambda("group1", loadGroupMomeryLamda)).
		AppendLambda(SetLanguage, compose.WithNodeName("a2")).
		AppendBranch(compose.NewChainBranch(branch2).AddChatTemplate("single2", chatTemplate).AddChatTemplate("group2", groupChatTemplate)).
		AppendLambda(agentLambda, compose.WithNodeKey("a3"))
	s.app, err = Chain.Compile(context.Background(), compose.WithGraphName("chat"))
	if err != nil {
		log.Println("编译Chain失败:", err)
	}

}

func (s *ChatChain) Stream(input map[string]any) (*schema.StreamReader[*schema.Message], error) {
	s.mu.Lock()
	// 如果存在旧的context，先取消它
	if s.currentCancel != nil {
		s.currentCancel()
	}
	s.currentCtx, s.currentCancel = context.WithCancel(context.Background())
	ctx := s.currentCtx
	ctx = context.WithValue(ctx, "session_id", input["session_id"])

	s.mu.Unlock()
	sessionID := input["session_id"].(int)
	msgID, ok := input["msg_id"].(int64)
	if !ok {
		logs.Logger.Error("没有提供有效的消息ID")
		msgID = 0
	}
	chatModelHandler := &callbacks2.ModelCallbackHandler{
		OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *model.CallbackInput) context.Context {
			fmt.Println("========[OnStart]model=========")
			fmt.Println("当前时间:", time.Now().Format("2006-01-02 15:04:05.000"))
			fmt.Println("runInfo:", runInfo)

			fmt.Println("model input:", input)
			return ctx
		},
		OnEnd: nil,
		OnEndWithStreamOutput: func(ctx context.Context, runInfo *callbacks.RunInfo, output *schema.StreamReader[*model.CallbackOutput]) context.Context {
			fmt.Println("========[OnEndStream]model=========")
			fmt.Println("当前时间:", time.Now().Format("2006-01-02 15:04:05.000"))
			fmt.Println(runInfo)
			//var graphInfoName = react.GraphName
			go func() {
				defer func() {
					if err := recover(); err != nil {
						fmt.Println("[OnEndStream] panic err:", err)
					}
				}()

				defer output.Close() // remember to close the stream in defer
				combinedMsg := ""
				completeMsg := ""
				tokenCount := 0

				var index int32
				//var err error
				index = 0
				//msgID = 0
				//msgID, err = dao.CreateEmptyMessage(dao.SqliteDB, sessionID, "assistant", s.roleName)
				//if err != nil {
				//	logs.Logger.Error("创建空消息失败", zap.Error(err))
				//}
				for {

					frame, err := output.Recv()
					if errors.Is(err, io.EOF) {
						if len(combinedMsg) > 1 {
							combinedMsg = FilterEmoji(combinedMsg)
							//fmt.Println("发送句子片段:", combinedMsg)
							//tts && 发送到ws连接
							atomic.AddInt32(&index, 1)
							s.ttsQueue <- TTSRequest{Text: combinedMsg, MessageID: msgID, Index: index, Sender: s.roleName}
							//go tts(combinedMsg, msgID, index)
						}
						atomic.AddInt32(&index, 1)
						s.ttsQueue <- TTSRequest{Text: "EOF", MessageID: msgID, Index: index, Sender: s.roleName}
						break
					}
					if err != nil {
						fmt.Printf("internal error: %s\n", err)
						return
					}
					//有工具就跳出，不返回工具信息
					if len(frame.Message.ToolCalls) > 0 {
						if len(combinedMsg) > 1 {
							combinedMsg = FilterEmoji(combinedMsg)
							//fmt.Println("发送句子片段:", combinedMsg)
							//tts && 发送到ws连接
							atomic.AddInt32(&index, 1)
							s.ttsQueue <- TTSRequest{Text: combinedMsg, MessageID: msgID, Index: index, Sender: s.roleName}
						}
						atomic.AddInt32(&index, 1)
						s.ttsQueue <- TTSRequest{Text: "EOF", MessageID: msgID, Index: index, Sender: s.roleName}
						break
					}

					if frame.TokenUsage == nil {
						tokenCount = 0
					} else {
						tokenCount = frame.TokenUsage.TotalTokens
					}

					combinedMsg += frame.Message.Content
					completeMsg += frame.Message.Content

					/// 分句处理，每次只提取一句
					sentence, remainder := SplitSentence(combinedMsg, 8, 60)
					if sentence != "" {
						//s.conn.WriteJSON(sentence)
						sentence = FilterEmoji(sentence)
						//fmt.Println("发送句子片段:", sentence)
						//tts && 发送到ws连接
						atomic.AddInt32(&index, 1)
						s.ttsQueue <- TTSRequest{Text: sentence, MessageID: msgID, Index: index, Sender: s.roleName}
						//go tts(sentence, msgID, index)
						// 更新combinedMsg为剩余未完成的部分
						combinedMsg = remainder
					}

				}
				logs.Logger.Debug("LLM完整回复", zap.String("content", completeMsg))
				if completeMsg != "" {

					go func() {
						_, err := SaveMessageToSession(dao.SqliteDB, sessionID, "assistant", s.roleName, completeMsg, tokenCount)
						if err != nil {
							logs.Logger.Error("保存ai消息失败", zap.Error(err))
						} else {
							logs.Logger.Info("保存ai消息成功", zap.String("content", completeMsg))
						}

					}()

				}
			}()
			return ctx
		},
		OnError: nil,
	}
	toolHandler := callbacks2.ToolCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *tool.CallbackInput) context.Context {
			fmt.Println("========[OnStart]tool=========")
			fmt.Println("当前时间:", time.Now().Format("2006-01-02 15:04:05.000"))
			fmt.Println("runInfo", info)
			fmt.Println("tool input:", input)
			return ctx
		},
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *tool.CallbackOutput) context.Context {
			fmt.Println("========[OnEnd]tool=========")
			fmt.Println("当前时间:", time.Now().Format("2006-01-02 15:04:05.000"))
			fmt.Println("runInfo", info)
			fmt.Println("tool output:", output)
			return ctx
		},
		OnEndWithStreamOutput: func(ctx context.Context, runInfo *callbacks.RunInfo, output *schema.StreamReader[*tool.CallbackOutput]) context.Context {
			fmt.Println("========[OnEndStream]tool=========")
			fmt.Println("当前时间:", time.Now().Format("2006-01-02 15:04:05.000"))
			fmt.Println(runInfo)
			//var graphInfoName = react.GraphName
			go func() {
				defer func() {
					if err := recover(); err != nil {
						fmt.Println("[OnEndStream] panic err:", err)
					}
				}()

				defer output.Close() // remember to close the stream in defer

				for {
					frame, err := output.Recv()
					if errors.Is(err, io.EOF) {
						// finish
						break
					}
					if err != nil {
						fmt.Printf("internal error: %s\n", err)
						return
					}

					s, err := json.Marshal(frame)
					if err != nil {
						fmt.Printf("internal error: %s\n", err)
						return
					}

					//if runInfo.Name == graphInfoName { // 仅打印 graph 的输出, 否则每个 stream 节点的输出都会打印一遍
					fmt.Printf("%s: %s\n", runInfo.Name, string(s))
					//}
				}

			}()
			return ctx
		},
		OnError: func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			fmt.Println("========[OnError]tool=========")
			fmt.Println("当前时间:", time.Now().Format("2006-01-02 15:04:05.000"))
			fmt.Println("runInfo", info)
			fmt.Println("tool error:", err)
			return ctx
		},
	}
	handler := callbacks2.NewHandlerHelper().ChatModel(chatModelHandler).Tool(&toolHandler).Handler()
	return s.app.Stream(ctx, input, compose.WithCallbacks(handler))

}
func (s *ChatChain) StopStream() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentCancel != nil {
		s.currentCancel()
		s.currentCancel = nil
	}
}
