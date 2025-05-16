package handler

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"virtugo/internal/dao"
)

// NewConversationRequest 新建会话请求结构体
type NewConversationRequest struct {
	Title string `json:"title"` // 可选，如果不提供则使用默认标题
}

// NewConversation 创建新会话
func NewConversation(c *gin.Context) {
	// 从上下文中获取用户名（由JWT中间件设置）
	username, exists := c.Get("username")
	if !exists {
		c.JSON(401, gin.H{
			"error": "未授权：用户未登录",
		})
		return
	}

	// 解析请求体
	var req NewConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果解析失败也没关系，会使用默认标题
	}

	db := dao.SqliteDB
	if db == nil {
		c.JSON(500, gin.H{
			"error": "数据库连接未初始化",
		})
		return
	}
	var err error
	// 创建新会话
	var sessionID int64
	if req.Title == "" {
		// 使用默认标题
		sessionID, err = dao.CreateSession(db, username.(string))
	} else {
		// 使用自定义标题
		sessionID, err = createSessionWithTitle(db, username.(string), req.Title)
	}

	if err != nil {
		c.JSON(500, gin.H{
			"error": "创建会话失败: " + err.Error(),
		})
		return
	}

	// 获取新创建的会话信息
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

	// 返回新创建的会话信息
	c.JSON(200, gin.H{
		"id":         sessionID,
		"title":      title,
		"created_at": createdAt,
		"updated_at": updatedAt,
	})
}

// createSessionWithTitle 创建带有自定义标题的会话
func createSessionWithTitle(db *sql.DB, username string, title string) (int64, error) {
	query := `INSERT INTO 会话聊天 (username, title) VALUES (?, ?)`
	result, err := db.Exec(query, username, title)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
