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
)

type ServiceContext struct {
	chatModel *openai.ChatModel //Todo后续会改进为多个模型
	Ctx       context.Context
	//Template  prompt.ChatTemplate
	Message []*schema.Message
	Agent   *react.Agent
	Chain   *compose.Chain[string, *schema.Message]
	app     compose.Runnable[string, *schema.Message]
}

func (s *ServiceContext) InitServiceContext() {
	s.Ctx = context.Background()
	s.chatModel, _ = openai.NewChatModel(s.Ctx, &openai.ChatModelConfig{
		Model:  "gpt-4o-mini", // 使用的模型版本
		APIKey: os.Getenv("OPENAI_API_KEY"),
	})
	persona := "你是一只猫娘,你需要用可爱的语气回答我"
	saveMemoryTool := GetSaveMemoryTool()
	toolsConfig := compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{saveMemoryTool},
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
	s.Chain = compose.NewChain[string, *schema.Message]()
	s.Chain.
		AppendLambda(loadMomeryLamda).
		AppendChatTemplate(prompt.FromMessages(schema.FString,
			// 系统消息模板
			schema.SystemMessage("1.你的回答应该尽量简短。2.积极地自主使用记忆存储工具同时不轻易储存已经在记忆里的内容。3.发言尽量贴近角色的设定"),
			schema.SystemMessage("长期记忆:{long_term_memory}"),
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
