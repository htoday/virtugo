# VirtuGo

## 📝 项目简介

VirtuGo 是一个功能强大的智能助手系统，基于 Go 语言Eino框架开发，整合了多种先进 AI 技术，为用户提供全方位的智能交互体验。系统设计轻量化，依赖最小化，同时保持强大功能和高性能表现。

## ✨ 核心特性

- **🔍 关键词唤醒**：可以设置使用关键词进行唤醒
- **🎙️ ASR 语音识别**：将语音输入转换为文本，支持实时交互
- **🔊 TTS 语音合成**：支持语音合成服务（Fish_audio、EdgeTTS），自然流畅的语音输出
- **🌐 多语言支持**：内置多语言处理能力，无缝切换不同语言环境
- **🧠 基于 RAG 的长期记忆**：利用检索增强生成技术构建持久化记忆系统
- **💾 嵌入式数据存储**：全部使用嵌入式数据库，无需外部依赖，部署简单高效
- **🔌 工具集成**：支持 MCP (Model Context Protocol) 工具扩展，如 Brave 搜索

## 🔧 与前端交互

![image](https://github.com/user-attachments/assets/7bb57439-ed6b-4bad-bc8a-f56b77397b6a)

  -可爱的 Live2D 模型: 以动态、生动的方式展示角色
  
  -完全可自定义: 支持修改外观、表情和动作
  
  -实时口型同步: 角色的嘴型会根据语音输出自动匹配

## 🚀 快速开始

### 安装要求

- Go 1.18+

目前该项目还没有稳定可运行的版本，我们会尽快推出！

### 配置说明

在 `config.yaml` 中设置：

```yaml
# 模型配置
model_info:
  base_url:
  api_key:
  model_name: "gpt-4o-mini"  # 使用的模型
  temperature: 0.8           # 创造性参数
  persona: "你是一个可爱积极的ai助理..."  # 助手人格设定

# 语音合成配置
tts:
  service_type: "edge"       # TTS 服务类型
  edge_tts_voice: "zh-CN-XiaoxiaoNeural"  # Edge TTS 语音

# 默认语言
language: "jp"  # 支持多种语言切换
```

### MCP 工具配置

在 `mcp.json` 中添加外部工具：

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

## 💡 使用场景

- **智能助手**：日常问答、任务规划、信息查询
- **多语言交流**：自动语言识别和翻译
- **知识管理**：长期记忆存储重要信息
- **语音交互**：无需键盘，直接语音对话

## 🔮 未来规划

- 重构长期记忆系统，构建用户画像，kv 异步存,同步调...有idea的话欢迎提issue
- 优化语音交互，支持更多asr&tts方案
- 优化前端的交互界面，等我买个cursor来写
- 强化多模态能力

---
