package handler

import (
	"github.com/gin-gonic/gin"
	"virtugo/internal/dao"
)

// ChangeUsernameRequest 请求结构体
type ChangeUsernameRequest struct {
	NewUsername string `json:"new_username" binding:"required"`
}

// ChangeUsername 修改当前用户的用户名
func ChangeUsername(c *gin.Context) {
	oldUser, exists := c.Get("username")
	if !exists {
		c.JSON(401, gin.H{"error": "未授权：用户未登录"})
		return
	}

	var req ChangeUsernameRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误：new_username 必填"})
		return
	}

	if err := dao.ChangeUsername(dao.SqliteDB, oldUser.(string), req.NewUsername); err != nil {
		c.JSON(500, gin.H{"error": "修改用户名失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "用户名修改成功"})
}
