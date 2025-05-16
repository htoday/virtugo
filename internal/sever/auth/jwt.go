package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("your_secret_jwt_key") // 请在生产环境中使用安全的密钥

// Claims 是 JWT 携带的自定义信息
type Claims struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// GenerateToken 生成访问令牌和刷新令牌
func GenerateToken(username, password string) (string, string, error) {
	accessToken, err := GenerateAccessToken(username, password)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := GenerateRefreshToken(username, password)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// GenerateAccessToken 生成短期访问令牌
func GenerateAccessToken(username, password string) (string, error) {

	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		Username:  username,
		Password:  password,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken 生成长期刷新令牌
func GenerateRefreshToken(username, password string) (string, error) {
	// 设置令牌有效期为30天
	expirationTime := time.Now().Add(30 * 24 * time.Hour)

	claims := &Claims{
		Username:  username,
		Password:  password,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析访问令牌
func ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("令牌已过期")
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, errors.New("无效的令牌签名")
		}
		return nil, errors.New("无法解析令牌: " + err.Error())
	}

	if !token.Valid {
		return nil, errors.New("无效的令牌")
	}

	// 验证令牌类型
	if claims.TokenType != "access" {
		return nil, errors.New("无效的令牌类型")
	}

	return claims, nil
}

// ParseRefreshToken 解析刷新令牌
func ParseRefreshToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("刷新令牌已过期")
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, errors.New("无效的刷新令牌签名")
		}
		return nil, errors.New("无法解析刷新令牌: " + err.Error())
	}

	if !token.Valid {
		return nil, errors.New("无效的刷新令牌")
	}

	// 验证令牌类型
	if claims.TokenType != "refresh" {
		return nil, errors.New("无效的令牌类型")
	}

	return claims, nil
}
