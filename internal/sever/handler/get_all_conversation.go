package handler

import (
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"virtugo/internal/dao"
)

// GetAllConversations 获取用户的所有会话信息
func GetAllConversations(c *gin.Context) {
	// 从上下文中获取用户名（由JWT中间件设置）
	username, exists := c.Get("username")
	if !exists {
		c.JSON(401, gin.H{
			"error": "未授权：用户未登录",
		})
		return
	}

	// 获取数据库连接
	db := dao.SqliteDB
	if db == nil {
		c.JSON(500, gin.H{
			"error": "数据库连接未初始化",
		})
		return
	}

	// 获取用户的所有会话
	sessions, err := dao.GetAllSessions(db, username.(string))
	if err != nil {
		c.JSON(500, gin.H{
			"error": "获取会话失败: " + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"conversations": sessions,
	})
}
