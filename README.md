# VirtuGo

## 📝 项目简介

VirtuGo 是一个开箱即用的，基于 Go 语言Eino框架开发的情感陪伴助手系统

整合多种先进 AI 技术与语音技术，赋予live2d模型有趣生动的灵魂，提供真实的对话和情感交互体验

得益于Go语言的特性，无需额外下载运行依赖，且轻量占用低

## ✨ 核心功能

🚀 **跨平台支持**：兼容 macOS 和 Windows，无需额外配置，开箱即用。

🧠 **广泛的模型支持**：支持接入多种主流大语言模型，轻松切换，满足不同场景需求。

🎤 **可实时打断的语音交互**：实现自然流畅的人机对话，支持用户随时打断并重新引导对话方向。

🧑‍🤝‍🧑 **多角色语音聊天室**：可同时与多个自定义 Live2D 虚拟角色聊天互动

💾 **聊天记录持久化**：随时切换到之前的对话

🧩 **MCP 扩展支持**：内置 MCP 拓展机制，可灵活拓展功能，快速集成自定义模块和能力。

## 🌐 与前端交互

前端库：https://github.com/aquamarine-z/vitrugo-frontend-next

<img width="1702" alt="前端图片" src="https://github.com/user-attachments/assets/a3470b47-44e8-447c-bb51-335ed16c7ebe" />


-完全可自定义: 支持修改为任意live2d模型

-实时口型同步: 角色的嘴型会根据语音输出自动匹配

## 🚀 快速开始

从release中下载对应的版本，解压后再填好配置信息

先访问localhost:8081/login.html注册账号

再访问localhost:8082/live.html就能使用啦

### 安装要求

- chrome/edge浏览器

### 开发要求

- go 1.24+

### 配置说明

前端配置在`out/live2d-models.json`中配置：

```json
[
  {
    "name": "小爱",
    "model": "mao_pro_zh/runtime/mao_pro.model3.json"
  },
  {
    "name": "日和",
    "model": "hiyori_pro_zh/runtime/hiyori_pro_t11.model3.json"
  }
]
```

后端配置在 `config.yaml` 中设置：

```yaml
llm模型配置
models:
  小爱: #重要的角色名，务必与前端的配置文件角色名保持一致
    model_info:
      api_type: "openai"  #openai表示使用支持openai格式的apikey，目前只写了openai和ollama两种格式的代码，不过eino支持的格式应该都能支持
      model_name: "" 
      role_name: "小爱" #角色名，务必跟上一级保持一致
      base_url: "" #平台对应的openai的baseurl 或者 ollama的baseurl
      api_key: "" 
      temperature: 0.8 #温度，温度越高随机性越强
      persona: "你是一个可爱积极的ai助理，名字是小爱，喜欢聊天,语气元气可爱活泼。" #ai角色设定
      system_prompt: "" #角色的提示词
    tts:
      service_type: "edge" #tts方式，目前只支持edge和fish_audio 欢迎来帮忙添加
      edge_tts_voice: "zh-CN-XiaoyiNeural" #edge的声线，可以去网上搜一下有哪些
      fish_audio_key: "" #fish_audio的key
      fish_audio_voice: "" #fish_audio网站内一个声音模型的网址的末尾可以看到这个参数，即模型ID
    tools:
      duckduckgo:
        is_enable: false #是否启用eino内置的duckduckgo联网搜索工具
      wikipedia:
        is_enable: false #是否启用eino内置的wikipedia搜索工具
        user_agent: "" 
      mcp_tool: #是否启用mcp
        is_enable: false
  第二个ai名: 略
    

history: 
  max_length: 20 #上下文保留长度，默认10
key_word_is_enable: false #是否启用关键词识别，启用后，打断和说话都必须带关键词
pre_generate_amount: 0 #提前生成量，默认1，群聊场景下提前生成n条消息，设为0的话一个ai说完话之后另一个ai才会开始生成（生成速度一般远快于播放速度）

backend_port: 8081 #后端占用端口
frontend_port: 8082 #前端占用端口
auth_key: "1145141919810" #注册用的key，可以随便改

```

### MCP 工具配置

在 `mcp.json` 中添加外部工具（以brave搜索为例）：

```json
{
  "mcpServers": {
    "brave-search": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "env": {
        "BRAVE_API_KEY": "您的API密钥"
      }
    }
  }
}
```

## 架构介绍

储存使用的sqlite，后续考虑引入持久化kv存储

多ai群聊用例的大致时序图（省略了vad和kws等）
![时序图](https://github.com/user-attachments/assets/8d3efb8f-310d-40ad-9794-5d1782b013eb)



部分模型如grok会在调用工具前回复一段文本，这在正常使用eino的react agent情况下会错误不进入tool node
本项目通过闭包将语音队列和ws连接传入callback，直接在callback里处理ai生成的消息来解决(如果有更好的方案欢迎提出👏)
因此你甚至可以提示词里写‘调用工具前给我说一声你在调用什么工具’

## 🔮 未来规划

- 多模态能力（图像）
- 优化群聊交互逻辑/优化主动对话逻辑，支持更多asr&tts方案
- 情感识别能力
- 长期记忆模块（考虑使用redis）
- 接入discord/qq等聊天平台
- 字幕/多语言字幕
- Live2d表情与动作控制
- Live2d更丰富的待机动作
- 桌面端应用

------

一点心得
开发期间从一些优秀的前辈开源项目（open-llm-vtuber,amadeus system等）学习了很多，包括设计模式和技术栈选型
在这里感谢开源开发者大佬无私的分享



