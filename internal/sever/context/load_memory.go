package context

import (
	"context"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"log"
	"virtugo/internal/dao"
)

func InitLoadMemory() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input string) (map[string]any, error) {

		collection, err := dao.ChromemDB.GetOrCreateCollection("memory", nil, nil)
		if err != nil {
			log.Println("获取memory集合失败", err)
		}
		count := collection.Count()
		var nResult int
		if count >= 5 {
			nResult = 5
		} else {
			nResult = count
		}
		result, err := collection.Query(ctx, input, nResult, nil, nil)
		if err != nil {
			log.Println("查询memory集合失败", err)
		}
		memories := ""
		for i, item := range result {
			memories += item.Content
			if i != len(result)-1 {
				memories += ","
			}
		}
		log.Println("查询到的memory:", memories)

		rows, err := dao.SqliteDB.Query("SELECT role, content, timestamp FROM messages ORDER BY timestamp DESC LIMIT 5")
		if err != nil {
			log.Println("获取上下文失败", err)
		}
		defer rows.Close()

		// 创建一个切片来存储消息
		var chatHistory []*schema.Message

		// 遍历查询结果
		for rows.Next() {
			var role, content, timestamp string
			err := rows.Scan(&role, &content, &timestamp)
			if err != nil {
				log.Fatal(err)
			}
			// 根据角色创建相应的消息实例
			var msg *schema.Message
			if role == "user" {
				msg = schema.UserMessage(content)
			} else if role == "ai" {
				msg = schema.AssistantMessage(content, nil)
			} else {
				log.Fatalf("未知的角色: %s", role)
			}
			// 将消息添加到 chatHistory 切片
			chatHistory = append(chatHistory, msg)
		}
		// 检查遍历过程中是否有错误
		if err := rows.Err(); err != nil {
			log.Println("转换过程遇到问题", err)
		}
		output := map[string]any{
			"long_term_memory": memories,
			"chat_history":     chatHistory,
			"question":         input,
		}
		go InsertMessage(dao.SqliteDB, "user", input)
		return output, nil
	})
}

//func BuildMap() *compose.Lambda {
//	return  compose.InvokableLambda(func(ctx context.Context, input string) (map[string]any, error) {
//		messageMap:=map[string]any{
//
//		}
//	}
//}

//func (s *ServiceContext) LoadMemory (ctx context.Context) {
//	collection, err := dao.ChromemDB.GetOrCreateCollection("memory", nil, nil)
//	if err != nil {
//		log.Println("获取memory集合失败", err)
//	}
//	collection.Query(ctx, s.Message.content)
//	return
//}
