package config

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/eino/components/tool"
	"os"
	"virtugo/internal/mcp"
	"virtugo/logs"
)

// MCPTool 表示一个MCP工具的信息
type MCPTool struct {
	Name    string   // 工具名称
	Command string   // 命令字符串 (如果是command模式)
	Args    []string // 命令参数 (如果是command模式)
	Env     []string // 环境变量，格式为KEY=VALUE
	URL     string   // SSE服务URL (如果是SSE模式)
	Type    string   // 工具类型："command" 或 "sse"
}

// MCPServerConfig 表示MCP服务器配置的不同可能格式
type MCPServerConfig struct {
	// Command模式字段
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`

	// SSE模式字段
	URL string `json:"url,omitempty"`
}

// MCPConfig 表示整个MCP配置
type MCPConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

var AllEinoTools []tool.BaseTool

func GetEinoTools() []tool.BaseTool {
	return AllEinoTools
}

func LoadMCPJson() {
	path := "mcp.json"
	ctx := context.Background()
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		logs.Logger.Warn("mcp文件不存在或路径不正确")
		return // 文件不存在
	}

	toolsInfo, err := ParseMCPJSON(path)
	if err != nil {
		logs.Logger.Error("加载mcp配置失败")
	}
	for i, k := range toolsInfo {
		logs.Logger.Info(string(i) + " 正在加载工具 " + k.Name)
		if k.Type == "studio" {
			tools := mcp.GetStudioTool(ctx, k.Command, k.Env, k.Args)
			AllEinoTools = append(AllEinoTools, tools...)
		}
		if k.Type == "sse" {
			tools := mcp.GetSSETool(ctx, k.URL)
			AllEinoTools = append(AllEinoTools, tools...)
		}
	}

}

// ParseMCPJSON 从文件路径解析MCP JSON，返回工具信息切片
func ParseMCPJSON(path string) ([]MCPTool, error) {
	// 读取JSON文件
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 解析JSON
	var config MCPConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// 创建工具信息切片
	tools := make([]MCPTool, 0, len(config.MCPServers))

	// 填充工具信息
	for name, serverInfo := range config.MCPServers {
		tool := MCPTool{
			Name: name,
		}

		// 确定工具类型并设置相应字段
		if serverInfo.URL != "" {
			// 这是SSE类型
			tool.Type = "sse"
			tool.URL = serverInfo.URL
		} else if serverInfo.Command != "" {
			// 这是studio类型
			tool.Type = "studio"
			tool.Command = serverInfo.Command
			tool.Args = serverInfo.Args
			tool.Env = make([]string, 0, len(serverInfo.Env))

			// 将环境变量映射转换为字符串切片，格式为 "KEY=VALUE"
			for key, value := range serverInfo.Env {
				tool.Env = append(tool.Env, key+"="+value)
			}
		}

		tools = append(tools, tool)
	}

	return tools, nil
}
