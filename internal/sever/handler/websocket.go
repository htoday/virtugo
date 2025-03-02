package handler

import (
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"virtugo/internal/sever/context"
	model "virtugo/internal/sever/message_model"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // 允许跨域
}

func HandleWebsocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket 升级失败:", err)
		return
	}
	defer conn.Close()
	serviceContext := context.ServiceContext{}
	serviceContext.InitServiceContext()
	fmt.Println("创建连接成功")
	for {
		// 读取消息
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("读取消息错误:", err)
			break
		}

		switch messageType {
		case websocket.TextMessage: // 文本消息
			fmt.Println("收到文本消息:", string(msg))
			var inputMessage model.TextMessage
			err := json.Unmarshal(msg, &inputMessage)
			if err != nil {
				log.Println("解析消息错误:", err)
			}
			switch inputMessage.Type {
			case "text":
				//err := serviceContext.AddUserQuestion(inputMessage.Content)
				//if err != nil {
				//	log.Println("添加用户问题失败:", err)
				//}
				streamResult, err := serviceContext.Stream(inputMessage.Content)
				if err != nil {
					log.Println("获取流失败:", err)
					continue
				}
				reportStream(streamResult, conn)
			case "image":
				fmt.Println("处理图片消息:", inputMessage.Content)
			default:
				fmt.Println("未知消息类型:", inputMessage.Type)
			}
			//conn.WriteMessage(websocket.TextMessage, []byte(response))

		case websocket.BinaryMessage: // 二进制消息
			log.Println("收到二进制消息，长度:", len(msg))
			conn.WriteMessage(websocket.TextMessage, []byte("服务器收到二进制消息"))

		default:
			log.Println("未知消息类型:", messageType)
		}
	}
}

func reportStream(sr *schema.StreamReader[*schema.Message], conn *websocket.Conn) {
	defer sr.Close()

	var combinedMessages string

	for {
		message, err := sr.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("recv failed: %v", err)
			return
		}
		// 将消息拼接到 combinedMessages
		combinedMessages += message.Content
		// 将消息发送到 WebSocket 连接
		err = conn.WriteJSON(message)
		if err != nil {
			log.Printf("write to websocket failed: %v", err)
			return
		}
	}
	go context.SaveContext("ai", combinedMessages)
	// 输出拼接后的消息到日志
	log.Printf("流式回复:\n%s", combinedMessages)
}
