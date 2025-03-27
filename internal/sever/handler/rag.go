package handler

import (
	"context"
	file "github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino/components/document"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/philippgille/chromem-go"
	"go.uber.org/zap"
	"virtugo/internal/dao"
	"virtugo/logs"
)

func LoadFile(c *gin.Context) {
	ctx := context.Background()

	// 初始化加载器
	loader, err := file.NewFileLoader(ctx, &file.FileLoaderConfig{
		UseNameAsID: true,
	})
	if err != nil {
		panic(err)
	}

	// 加载文档
	docs, err := loader.Load(ctx, document.Source{
		URI: "./documents/atri.txt",
	})
	if err != nil {
		panic(err)
	}
	// 初始化分割器
	splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   50,
		OverlapSize: 5,
		LenFunc: func(s string) int {
			// eg: 使用 unicode 字符数而不是字节数
			return len([]rune(s))
		},
		Separators: []string{"\n", ",", "，", ".", "。", "?", "!"},
		KeepType:   recursive.KeepTypeEnd,
	})
	if err != nil {
		panic(err)
	}
	// 使用文档内容
	//for _, doc := range docs {
	//	//println(doc.Content)
	//
	//}
	// 执行分割
	results, err := splitter.Transform(ctx, docs)
	if err != nil {
		panic(err)
	}
	collection, err := dao.ChromemDB.GetOrCreateCollection("rag", nil, nil)
	// 处理分割结果
	for i, doc := range results {
		println("片段", i+1, ":", doc.Content)
		doc1 := chromem.Document{
			ID:      "rag" + uuid.New().String(),
			Content: doc.Content,
			Metadata: map[string]string{
				"source": "文档",
			},
		}
		err = collection.AddDocument(ctx, doc1)
		if err != nil {
			logs.Logger.Error("添加文档失败", zap.Error(err))
		}
	}
	logs.Logger.Info("文档加载完毕")
}
