package mcp

import (
	"context"
	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"virtugo/logs"
)

func GetSSETool(ctx context.Context, url string) []tool.BaseTool {
	cli, err := client.NewSSEMCPClient(url)
	if err != nil {
		logs.Logger.Error("mcp加载出错")
		return nil
	}
	err = cli.Start(ctx)
	if err != nil {
		logs.Logger.Error("mcp启动出错")
		return nil
	}
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "example-client",
		Version: "1.0.0",
	}

	_, err = cli.Initialize(ctx, initRequest)
	if err != nil {
		logs.Logger.Error("mcp初始化出错")
		return nil
	}
	//res, err := cli.ListTools(ctx, mcp.ListToolsRequest{})

	tools, err := mcpp.GetTools(ctx, &mcpp.Config{Cli: cli})
	if err != nil {
		logs.Logger.Error("eino转换mcp出错")
		return nil
	}
	return tools
}
