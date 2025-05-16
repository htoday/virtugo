package handler

import (
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"strconv"
	"virtugo/internal/dao"
)

// DeleteConversation 删除会话
func DeleteConversation(c *gin.Context) {
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

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		c.JSON(500, gin.H{
			"error": "开始事务失败: " + err.Error(),
		})
		return
	}

	// 首先删除会话相关的所有消息
	_, err = tx.Exec("DELETE FROM 消息 WHERE conversation_id = ?", sessionID)
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{
			"error": "删除会话消息失败: " + err.Error(),
		})
		return
	}

	// 然后删除会话本身
	_, err = tx.Exec("DELETE FROM 会话聊天 WHERE id = ?", sessionID)
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{
			"error": "删除会话失败: " + err.Error(),
		})
		return
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		c.JSON(500, gin.H{
			"error": "提交事务失败: " + err.Error(),
		})
		return
	}

	// 返回成功消息
	c.JSON(200, gin.H{
		"message": "会话删除成功",
		"id":      sessionID,
	})
}
