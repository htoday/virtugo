package middleware

import (
	"github.com/gin-gonic/gin"
	"strings"
	"time"
	"virtugo/internal/sever/auth"
)

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Refresh-Token")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-New-Access-Token")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先级: 请求头 > Cookie > URL查询参数
		tokenString := extractToken(c)

		if tokenString == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "未授权：缺少认证信息",
			})
			return
		}

		// 解析和验证Token
		claims, err := auth.ParseToken(tokenString)
		if err != nil {
			// 如果是access token过期，尝试使用refresh token
			if err.Error() == "令牌已过期" {
				refreshToken, err := extractRefreshToken(c)
				if err != nil || refreshToken == "" {
					c.AbortWithStatusJSON(401, gin.H{
						"error": "访问令牌已过期，请重新登录",
						"code":  "token_expired",
					})
					return
				}

				// 验证refresh token
				refreshClaims, err := auth.ParseRefreshToken(refreshToken)
				if err != nil {
					c.AbortWithStatusJSON(401, gin.H{
						"error": "刷新令牌无效：" + err.Error(),
						"code":  "invalid_refresh_token",
					})
					return
				}

				// 生成新的access token
				newAccessToken, err := auth.GenerateAccessToken(refreshClaims.Username, refreshClaims.Password)
				if err != nil {
					c.AbortWithStatusJSON(500, gin.H{
						"error": "生成新令牌失败：" + err.Error(),
					})
					return
				}

				// 在响应头中返回新的access token
				c.Header("X-New-Access-Token", newAccessToken)

				// 设置新的cookie
				c.SetCookie(
					"access_token",
					newAccessToken,
					int(24*time.Hour.Seconds()),
					"/",
					"",
					false,
					true,
				)

				// 将用户信息存储在上下文中
				c.Set("username", refreshClaims.Username)
				c.Next()
				return
			}

			c.AbortWithStatusJSON(401, gin.H{
				"error": "未授权：" + err.Error(),
				"code":  "invalid_token",
			})
			return
		}

		// 将用户信息存储在上下文中，以便后续使用
		c.Set("username", claims.Username)
		c.Next()
	}
}

// extractToken 从多个来源提取access token
func extractToken(c *gin.Context) string {
	// 1. 从Authorization头中获取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// 2. 从Cookie中获取
	token, err := c.Cookie("access_token")
	if err == nil && token != "" {
		return token
	}

	// 3. 从URL查询参数获取
	return c.Query("token")
}

// extractRefreshToken 提取refresh token
func extractRefreshToken(c *gin.Context) (string, error) {
	// 1. 从Cookie中获取
	token, err := c.Cookie("refresh_token")
	if err == nil && token != "" {
		return token, nil
	}

	// 2. 从请求头获取
	refreshHeader := c.GetHeader("X-Refresh-Token")
	if refreshHeader != "" {
		return refreshHeader, nil
	}

	// 3. 从URL查询参数获取
	token = c.Query("refresh_token")
	if token != "" {
		return token, nil
	}

	return "", nil
}
