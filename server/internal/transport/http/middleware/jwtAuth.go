package middleware

import (
	"vehivle/pkg/response"

	"github.com/gin-gonic/gin"
	"vehivle/pkg/jwt"
	"errors"
)

const (
	CookieAccessToken  = "access_token"
	CtxKeyUserID       = "current_user_id"
	CtxKeyUsername     = "current_username"
	CtxKeyRole         = "current_role"
)

func JWTAuth(secret string) gin.HandlerFunc{
	return func(c *gin.Context) {
		tokenString, err := c.Cookie(CookieAccessToken)
		if err != nil || tokenString == "" {
			response.FailAuth(c, "token 不存在")
			c.Abort()
			return 
		}
		claims, err := jwt.ParseToken(secret, tokenString)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				response.FailAuth(c, err.Error())
				c.Abort()
				return 
			}
			response.FailAuth(c, err.Error())
			c.Abort()
			return 
		}
		// set: 是gin.Context中的一个方法，用于设置上下文中的值
		// key可以随意定义，但是需要保证唯一性
		c.Set(CtxKeyUsername, claims.Username)
		c.Set(CtxKeyRole, claims.Role)
		c.Set(CtxKeyUserID, claims.UserID)
		c.Next()
	}
}