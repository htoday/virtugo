package context

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"log"
	"os"
	"sync/atomic"
	"virtugo/internal/config"
)

type ServiceContext struct {
	chatModel *openai.ChatModel //Todo后续会改进为多个模型
	Ctx       context.Context
	//Template  prompt.ChatTemplate
	Message []*schema.Message
	Agent   *react.Agent
	Chain   *compose.Chain[string, *schema.Message]
	app     compose.Runnable[string, *schema.Message]

	WorkStage string
	stopFlag  atomic.Bool
}

func (s *ServiceContext) InitServiceContext() {
	s.WorkStage = "sleeping"
	s.Ctx = context.Background()
	s.chatModel, _ = openai.NewChatModel(s.Ctx, &openai.ChatModelConfig{
		//Model:  "gpt-4o-mini", // 使用的模型版本
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		BaseURL: config.Cfg.ModelInfo.BaseUrl,
		Model:   config.Cfg.ModelInfo.ModelName,

		Temperature: &config.Cfg.ModelInfo.Temperature,
	})
	persona := config.Cfg.ModelInfo.Persona
	saveMemoryTool := GetSaveMemoryTool()
	//queryTextTool := GetQueryTextTool()
	changeLanguageTool := ChangeLanguage()
	tools := []tool.BaseTool{}
	tools = append(tools, saveMemoryTool)
	tools = append(tools, changeLanguageTool)

	mcpTools := config.GetEinoTools()
	tools = append(tools, mcpTools...)
	toolsConfig := compose.ToolsNodeConfig{
		Tools: tools,
	}
	var err error
	s.Agent, err = react.NewAgent(s.Ctx, &react.AgentConfig{
		Model:           s.chatModel,
		ToolsConfig:     toolsConfig,
		MessageModifier: react.NewPersonaModifier(persona),
	})
	if err != nil {
		log.Println("创建Agent失败:", err)
	}

	agentLambda, _ := compose.AnyLambda(s.Agent.Generate, s.Agent.Stream, nil, nil)
	loadMomeryLamda := InitLoadMemory()
	SetLanguage := SetLanguage()
	s.Chain = compose.NewChain[string, *schema.Message]()
	s.Chain.
		AppendLambda(loadMomeryLamda).
		AppendLambda(SetLanguage).
		AppendChatTemplate(prompt.FromMessages(schema.FString,
			// 系统消息模板
			schema.SystemMessage("现在的时间是: {current_time}"),
			schema.SystemMessage("1.你的回答应该尽量简短。2.你的回答控制在30字以内，直接用文字回答，模仿人类说话的情景，不要有任何口语表达中不会出现的符号。3.通过工具获取到的内容需要自己思考之后概括总结，不要原样返回消息。4.你需要热心的回答用户的问题，照顾他的感情，不要重复用户的话，不要冷漠。5.回答中不要有换行符号"),
			schema.SystemMessage("长期记忆:{long_term_memory}"),
			schema.SystemMessage("如果回答有多句话，用|分隔，例如 first sentence|second sentence——第一句话|第二句话"),
			schema.SystemMessage("{language_config}"),
			// 插入需要的对话历史（新对话的话这里不填）
			schema.MessagesPlaceholder("chat_history", true),

			// 用户消息模板
			schema.UserMessage("问题: {question}"),
		)).
		AppendLambda(agentLambda)
	s.app, err = s.Chain.Compile(context.Background())
	if err != nil {
		log.Println("编译Chain失败:", err)
	}

}

func (s *ServiceContext) Stream(input string) (*schema.StreamReader[*schema.Message], error) {
	//return s.Agent.Stream(s.Ctx, s.Message)
	//return s.chatModel.Stream(s.Ctx, s.Message)
	return s.app.Stream(s.Ctx, input)

}
func (s *ServiceContext) Generate() (*schema.Message, error) {
	return s.chatModel.Generate(s.Ctx, s.Message)

}
