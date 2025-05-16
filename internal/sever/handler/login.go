package handler

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
	"virtugo/internal/dao"
	"virtugo/internal/sever/auth"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required" form:"username"`
	Password string `json:"password" binding:"required" form:"password"`
}

func Login(c *gin.Context) {
	var loginReq LoginRequest
	if err := c.ShouldBind(&loginReq); err != nil {
		c.JSON(400, gin.H{
			"error": "请求参数错误" + err.Error(),
		})
		return
	}
	if loginReq.Username == "" || loginReq.Password == "" {
		c.JSON(400, gin.H{
			"error": "用户名或密码不能为空",
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

	// 从数据库中获取用户信息
	var storedPassword string
	err := db.QueryRow("SELECT password FROM 用户 WHERE username = ?", loginReq.Username).Scan(&storedPassword)

	// 处理用户不存在的情况
	if err == sql.ErrNoRows {
		c.JSON(401, gin.H{
			"error": "用户名或密码错误",
		})
		return
	}

	// 处理数据库查询错误
	if err != nil {
		c.JSON(500, gin.H{
			"error": "查询用户信息失败: " + err.Error(),
		})
		return
	}

	// 验证密码是否匹配
	if storedPassword != loginReq.Password {
		c.JSON(401, gin.H{
			"error": "用户名或密码错误",
		})
		return
	}

	// 生成访问令牌和刷新令牌
	accessToken, refreshToken, err := auth.GenerateToken(loginReq.Username, loginReq.Password)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "生成令牌失败: " + err.Error(),
		})
		return
	}

	// 设置 Cookie
	// 使用自定义header设置Cookie，添加SameSite=None
	c.Writer.Header().Add("Set-Cookie",
		fmt.Sprintf("access_token=%s; Max-Age=%d; Path=/; HttpOnly; SameSite=None; Secure",
			accessToken,
			int(24*time.Hour.Seconds())))

	c.Writer.Header().Add("Set-Cookie",
		fmt.Sprintf("refresh_token=%s; Max-Age=%d; Path=/; HttpOnly; SameSite=None; Secure",
			refreshToken,
			int(30*24*time.Hour.Seconds())))

	c.JSON(200, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    3600 * 24, // 24小时的秒数
		"token_type":    "Bearer",
	})
}
