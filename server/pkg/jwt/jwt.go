package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	issuer = "vehivle-admin"
)

var (
	ErrTokenExpired = errors.New("token 过期")
	ErrTokenInvalid = errors.New("token 无效")
)

// Claims 结构体声明
type Claims struct {
	UserID               string `json:"user_id"`
	Username             string `json:"username"`
	Role                 string `json:"role"`
	jwt.RegisteredClaims        // 注册声明
}

/**
* GenerateToken 生成 Token
* @param secret 密钥
* @param duration 持续时间
* @param userID 用户ID
* @param username 用户名
* @param role 角色
* @return string Token
* @return error 错误
 */
func GenerateToken(secret string, duration time.Duration, userID string, username string, role string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),               // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),               // 生效时间
			Issuer:    issuer,                                       // 签发者
		},
	}
	// 创建新的令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 签名令牌
	return token.SignedString([]byte(secret))
}

/**
* ParseToken 解析 Token
* @param token Token
* @param secret 密钥
* @return *Claims 声明
* @return error 错误
 */
func ParseToken(secret string, tokenString string) (*Claims, error) {
	claims := &Claims{}
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 检查签名方法
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			// 不支持的签名方法
			return nil, fmt.Errorf("%w: unsupported signing method", ErrTokenInvalid)
		}
		// 返回密钥
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenInvalid
		}
		return nil, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}
	if !token.Valid {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}
