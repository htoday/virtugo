package handler

import (
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"strconv"
	"virtugo/internal/dao"
)

// GetSessionMessagesByID 获取特定会话的所有消息
func GetSessionMessagesByID(c *gin.Context) {
	// 从上下文中获取用户名（由JWT中间件设置）
	username, exists := c.Get("username")
	if !exists {
		c.JSON(401, gin.H{
			"error": "未授权：用户未登录",
		})
		return
	}

	// 获取会话ID参数
	sessionIDStr := c.Param("id")
	sessionID, err := strconv.Atoi(sessionIDStr)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "无效的会话ID",
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

	// 验证会话是否属于当前用户
	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM 会话聊天 
		WHERE id = ? AND username = ?
	`, sessionID, username.(string)).Scan(&count)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "查询会话失败: " + err.Error(),
		})
		return
	}

	if count == 0 {
		c.JSON(403, gin.H{
			"error": "无权访问此会话或会话不存在",
		})
		return
	}

	// 获取会话的所有消息（不限制数量）
	messages, err := dao.GetSessionMessages(db, sessionID, 1000) // 设置一个较大的限制值
	if err != nil {
		c.JSON(500, gin.H{
			"error": "获取消息失败: " + err.Error(),
		})
		return
	}

	// 获取会话信息
	var title string
	var createdAt, updatedAt string
	err = db.QueryRow(`
		SELECT title, created_at, updated_at FROM 会话聊天 
		WHERE id = ?
	`, sessionID).Scan(&title, &createdAt, &updatedAt)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "获取会话信息失败: " + err.Error(),
		})
		return
	}

	// 返回会话信息和消息列表
	c.JSON(200, gin.H{
		"id":         sessionID,
		"title":      title,
		"created_at": createdAt,
		"updated_at": updatedAt,
		"messages":   messages,
	})
}
