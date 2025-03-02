package sever

import (
	"github.com/gin-gonic/gin"
	"virtugo/internal/sever/handler"
	"virtugo/internal/sever/middleware"
)

func StartSever(addr string, port string) {
	r := gin.New()

	v1 := r.Group("")
	v1.Use(middleware.CorsMiddleware()) // Add CORS middleware

	v1.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	v1.GET("/ws", handler.HandleWebsocket)

	r.Run(addr + ":" + port)
}
