package handler

import (
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"strconv"
	"virtugo/internal/dao"
)

// RenameConversationRequest 重命名会话请求结构体
type RenameConversationRequest struct {
	Title string `json:"title" binding:"required" form:"title"` // 新标题
}

// RenameConversation 重命名会话
func RenameConversation(c *gin.Context) {
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

	// 解析请求体
	var req RenameConversationRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

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

	// 更新会话标题
	err = dao.RenameSession(db, sessionID, req.Title)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "重命名会话失败: " + err.Error(),
		})
		return
	}

	// 获取更新后的会话信息
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

	// 返回更新后的会话信息
	c.JSON(200, gin.H{
		"id":         sessionID,
		"title":      title,
		"created_at": createdAt,
		"updated_at": updatedAt,
	})
}
