package context

import (
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func NewChatTemplate() *prompt.DefaultChatTemplate {
	template := prompt.FromMessages(schema.FString,
		// 系统消息模板
		schema.SystemMessage("现在的时间是: {current_time}"),
		schema.SystemMessage("这是一个语音通话的场景，用户的错别字请猜测理解，直接用文字回答，模仿人类说话的情景，不要有口语表达中不会出现的符号"),
		//schema.SystemMessage("如果需要调用 tool,如果需要调用 tool，直接输出 tool，不要输出文本"),
		schema.SystemMessage("如果需要调用tool,最好先告诉用户你会做什么，然后在中途调用tool call，减少用户等待时间"),
		schema.SystemMessage("从互联网获取的内容需要自己思考之后概括总结，不要原样返回消息和链接"),
		schema.SystemMessage("如果回复较长，分句用｜分隔开来"),
		//schema.SystemMessage("1.你的回答应该尽量简短。2.你的回答控制在30字以内，直接用文字回答，模仿人类说话的情景，不要有任何口语表达中不会出现的符号。3.通过工具获取到的内容需要自己思考之后概括总结，不要原样返回消息。4.你需要热心的回答用户的问题，照顾他的感情，不要重复用户的话，不要冷漠。5.回答中不要有换行符号"),
		schema.SystemMessage("长期记忆:{long_term_memory}"),
		//schema.SystemMessage("在用——分隔两种语言的回答，例如 first sentence——第一句话"),
		//schema.SystemMessage("如果回答有多句话，用|分隔，例如 first sentence|second sentence——第一句话|第二句话"),
		schema.SystemMessage("{language_config}"),
		// 插入需要的对话历史（新对话的话这里不填）
		schema.MessagesPlaceholder("chat_history", true),

		// 用户消息模板
		schema.UserMessage("问题: {question}"))
	return template
}
