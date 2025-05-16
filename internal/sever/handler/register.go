package handler

import (
	"github.com/gin-gonic/gin"
	"virtugo/internal/config"
	"virtugo/internal/dao"
)

// RegisterRequest 注册请求结构体
type RegisterRequest struct {
	Username string `json:"username" binding:"required" form:"username"`
	Password string `json:"password" form:"password"`
	AuthKey  string `json:"auth_key" binding:"required" form:"auth_key"`
}

// Register 处理用户注册
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{
			"error": "请求参数错误",
		})
		return
	}

	// 验证auth_key是否匹配配置中的值
	if req.AuthKey != config.Cfg.AuthKey {
		c.JSON(403, gin.H{
			"error": "管理员密钥无效",
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

	// 检查用户名是否已存在
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM 用户 WHERE username = ?", req.Username).Scan(&count)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "查询用户失败: " + err.Error(),
		})
		return
	}

	if count > 0 {
		c.JSON(400, gin.H{
			"error": "用户名已存在",
		})
		return
	}

	// 处理密码为空的情况
	password := req.Password
	if password == "" {
		password = "114514" // 设置默认密码
	}

	//// 密码加密
	//hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	//if err != nil {
	//	c.JSON(500, gin.H{
	//		"error": "密码加密失败",
	//	})
	//	return
	//}

	// 创建用户
	err = dao.CreateUser(db, req.Username, string(password), "")
	if err != nil {
		c.JSON(500, gin.H{
			"error": "创建用户失败: " + err.Error(),
		})
		return
	}

	// 为新用户创建一个默认会话
	sessionID, err := dao.CreateSession(db, req.Username)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "创建默认会话失败: " + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message":    "注册成功",
		"username":   req.Username,
		"session_id": sessionID,
	})
}
