package context

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
	"virtugo/internal/config"
	"virtugo/internal/dao"
	"virtugo/logs"
)

type saveMemoryReq struct {
	KeyInfo string `json:"key_info" jsonschema:"description=key information for conversation"`
}

type saveMemoryResp struct {
	status string
}

type queryTextReq struct {
	KeyInfo string `json:"key_info" jsonschema:"description=你想要查询的内容"`
}
type queryTextResp struct {
	text string
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

func ChangeLanguage() tool.InvokableTool {
	saveMemoryTool := utils.NewTool(
		&schema.ToolInfo{
			Name: "change_language",
			Desc: "切换你回复的语言,zh是中文，jp是日语,en是英语",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"language": {
					Type:     schema.String,
					Desc:     "你想要切换的语言,在zh,jp,en中选择",
					Required: true,
				},
			}),
		}, func(ctx context.Context, input *changeLanguageReq) (output *changeLanguageResp, err error) {
			resp := ""
			switch input.Language {
			case "zh":
				logs.Logger.Info("ai切换了语言为中文")
				config.Cfg.Language = "zh"
				resp = "切换完成"
			case "jp":
				logs.Logger.Info("ai切换了语言为日语")
				config.Cfg.Language = "jp"
				resp = "切换完成，请严格遵守格式，你需要用把回答翻译成不同的语言,你的回答需要先用日语，再用中文回答，中间用——分隔，例如 こんにちは——你好"
			case "en":
				logs.Logger.Info("ai切换了语言为英语")
				config.Cfg.Language = "en"
				resp = "切换完成，请严格遵守格式，你需要用把回答翻译成不同的语言,你的回答需要先用英语，再用中文回答，中间用——分隔，例如 hello——你好"
			}

			return &changeLanguageResp{status: resp}, nil
		},
	)
	return saveMemoryTool
}

func GetQueryTextTool() tool.InvokableTool {
	queryTextTool := utils.NewTool(
		&schema.ToolInfo{
			Name: "query_text",
			Desc: "从角色台词中查询相关的语句",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"key_info": {
					Type:     schema.String,
					Desc:     "你想要查询的内容，会根据语意相似性返回",
					Required: true,
				},
			}),
		}, func(ctx context.Context, input *queryTextReq) (output *queryTextResp, err error) {
			logs.Logger.Info("ai发起查询: " + input.KeyInfo)
			collection := dao.ChromemDB.GetCollection("rag", nil)
			count := collection.Count()
			var nResult int
			if count >= 3 {
				nResult = 3
			} else {
				if count == 0 {
					nResult = 1
				} else {
					nResult = count
				}
			}
			result, err := collection.Query(ctx, input.KeyInfo, nResult, nil, nil)
			if err != nil {
				logs.Logger.Error("获取rag集合失败", zap.Error(err))
			}
			ragContent := ""
			for i, item := range result {
				ragContent += item.Content
				if i != len(result)-1 {
					ragContent += ","
				}
			}
			logs.Logger.Info("查询到的资料内容:" + ragContent)
			return &queryTextResp{text: ragContent}, nil
		},
	)
	return queryTextTool
}
