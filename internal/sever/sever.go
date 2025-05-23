package sever

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"virtugo/internal/config"
	"virtugo/internal/sever/handler"
	"virtugo/internal/sever/handler/websocket"
	"virtugo/internal/sever/middleware"
)

func StartSever(addr string, port string) {
	r := gin.New()
	//全局cors中间件
	r.Use(middleware.CorsMiddleware())
	r.POST("/login", handler.Login)
	r.POST("/register", handler.Register)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Use(middleware.CorsMiddleware())
	v1 := r.Group("")
	v1.Use(middleware.JWTAuthMiddleware())
	//v1.Use(middleware.AuthMiddleware()) // Add logger middleware

	v1.GET("/ws", websocket.HandleWebsocket)
	v1.GET("/setting", handler.GetSetting)
	v1.GET("/conversation", handler.GetAllConversations)
	v1.GET("/conversation/:id", handler.GetSessionMessagesByID)
	v1.POST("/conversation", handler.NewConversation)
	v1.PUT("/conversation/:id", handler.RenameConversation)
	v1.DELETE("/conversation/:id", handler.DeleteConversation)
	v1.POST("/changeusername", handler.ChangeUsername)
	v1.POST("/load", handler.LoadFile)
	v1.GET("/user", handler.GetUserInfo)
	// 设置HTML模板
	//r.LoadHTMLGlob("templates/*")

	r1 := gin.New()
	r1.Static("/", "./out")
	go r1.Run(":" + strconv.Itoa(config.Cfg.FrontendPort))

	r.Run(addr + ":" + port)
}
