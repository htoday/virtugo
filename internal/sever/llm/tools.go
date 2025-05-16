package llm

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/philippgille/chromem-go"
	"go.uber.org/zap"
	"time"
	"virtugo/internal/dao"
	"virtugo/logs"
)

type saveMemoryReq struct {
	KeyInfo string `json:"key_info" jsonschema:"description=key information for conversation"`
}

type saveMemoryResp struct {
	status string
}

type sleepReq struct {
}
type sleepResp struct {
	status string
}

type changeLanguageReq struct {
	Language string `json:"language" jsonschema:"description=你想要切换的语言"`
}
type changeLanguageResp struct {
	status string
}

func GetSaveMemoryTool() tool.InvokableTool {
	saveMemoryTool := utils.NewTool(
		&schema.ToolInfo{
			Name: "save_memory",
			Desc: "提炼出与用户的重要回忆存入向量数据库",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"key_info": {
					Type:     schema.String,
					Desc:     "关于对话的关键信息",
					Required: true,
				},
			}),
		}, func(ctx context.Context, input *saveMemoryReq) (output *saveMemoryResp, err error) {
			// 获取当前时间
			currentTime := time.Now()

			// 获取时间戳（秒级别）
			timestamp := currentTime.Unix()

			// 将时间戳转换为字符串
			timestampStr := fmt.Sprintf("%d", timestamp)
			doc := chromem.Document{
				ID:      uuid.New().String(),
				Content: input.KeyInfo,
				Metadata: map[string]string{
					"timestamp": timestampStr,
					"cate":      "长期记忆",
				},
			}
			collection := dao.ChromemDB.GetCollection("memory", nil)
			err = collection.AddDocument(ctx, doc)
			if err != nil {
				logs.Logger.Error("添加文档失败", zap.Error(err))
				return &saveMemoryResp{status: "fail to save"}, err
			}
			logs.Logger.Info("ai调用了save工具,储存了：" + input.KeyInfo)
			return &saveMemoryResp{status: "ok"}, nil
		},
	)
	return saveMemoryTool
}

func Sleep() tool.InvokableTool {
	sleepTool := utils.NewTool(
		&schema.ToolInfo{
			Name: "go_to_sleep",
			Desc: "当用户明确表示出让ai休眠或说再见时，调用该工具进入休眠状态",
		}, func(ctx context.Context, input any) (output *sleepResp, err error) {
			logs.Logger.Info("调用了休眠工具ai进入休眠状态")
			return &sleepResp{status: ""}, nil
		},
	)
	return sleepTool
}
