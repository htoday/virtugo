package handler

import "github.com/gin-gonic/gin"

func GetUserInfo(c *gin.Context) {
	// 从上下文中获取用户名
	username, exists := c.Get("username")
	if !exists {
		c.JSON(401, gin.H{"error": "未授权：用户未登录"})
		return
	}

	// 返回用户名
	c.JSON(200, gin.H{"username": username})

}
