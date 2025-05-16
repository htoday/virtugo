package llm

import (
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func NewChatTemplate() *prompt.DefaultChatTemplate {
	template := prompt.FromMessages(schema.FString,
		//历史消息，一定要放最上面，不然系统消息ai会忽视，我也不知道为什么
		schema.MessagesPlaceholder("chat_history", true),
		// 系统消息模板
		schema.SystemMessage("现在的时间是: {current_time}"),
		schema.SystemMessage("这是一个语音聊天室场景,房间内的人类参与者有{username},房间内的ai参与者有{group_member}"),
		schema.SystemMessage("这是一个语音通话的场景，用户的错别字请猜测理解，你的发言应该简洁，不要啰嗦重复，模仿人类说话的情景，不要有口语表达中不会出现的符号"),
		schema.SystemMessage("不要使用markdown的格式输出，不要使用`\n`,不要使用emoji表情,,直接输出文本"),
		schema.SystemMessage("你会看到json格式的历史会话，但是不要使用json格式输出，请直接输出文本"),
		//schema.SystemMessage("如果需要调用 tool,如果需要调用 tool，直接输出 tool，不要输出文本"),
		//schema.SystemMessage("如果需要调用tool,最好先告诉用户你会做什么，然后在中途调用tool call，减少用户等待时间。"),
		schema.SystemMessage("从互联网获取的内容需要自己思考之后概括总结，不要原样返回消息和链接"),
		schema.SystemMessage("{prompt}"),

		// 用户消息模板
		schema.UserMessage("问题: {question}"))
	return template
}

func NewGroupChatTemplate() *prompt.DefaultChatTemplate {
	template := prompt.FromMessages(schema.FString,
		schema.MessagesPlaceholder("chat_history", true),
		// 系统消息模板
		schema.SystemMessage("现在的时间是: {current_time}"),
		schema.SystemMessage("这是一个语音聊天室场景,房间内的人类参与者有{username},房间内的ai参与者有{group_member},你需要关注最新的一条消息，请交替使用长短不一的响应，使对话更具互动性和趣味性，你可以与其他AI参与者互动，并在对话中自由探索有趣的话题，避免使用冒号:让对话更加自然"),
		schema.SystemMessage("同时这是一个语音通话的场景，用户的错别字请猜测理解，你的发言应该简洁，不要长篇大论，模仿真实人类聊天的情景，不要有口语表达中不会出现的符号"),
		schema.SystemMessage("不要使用markdown的格式输出，不要使用`\n`,不要使用emoji表情"),
		schema.SystemMessage("你会看到json格式的历史会话，但是不要使用json格式输出，请直接输出文本"),
		schema.SystemMessage("从互联网获取的内容需要自己思考之后概括总结，不要原样返回消息和链接"),
		schema.SystemMessage("{prompt}"),
		schema.UserMessage("根据上下文参与群聊"))

	return template
}
