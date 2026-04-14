package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"

	"vehivle/pkg/jwt"
	"vehivle/pkg/response"
)

const (
	CookieAccessToken = "access_token"
	CtxKeyUserID      = "current_user_id"
	CtxKeyUsername    = "current_username"
	CtxKeyRole        = "current_role"
)

// JWTAuth 从 httpOnly Cookie 读取 Access Token，解析 Claims 并注入 Context。
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie(CookieAccessToken)
		if err != nil || tokenString == "" {
			response.FailAuth(c, "缺少有效的认证令牌")
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(secret, tokenString)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				response.FailAuth(c, "认证令牌已过期")
				c.Abort()
				return
			}
			response.FailAuth(c, "认证令牌无效")
			c.Abort()
			return
		}

		c.Set(CtxKeyUserID, claims.UserID)
		c.Set(CtxKeyUsername, claims.Username)
		c.Set(CtxKeyRole, claims.Role)
		c.Next()
	}
}

// RequireRole 角色鉴权中间件，仅允许指定角色通过。
// 用法：router.Use(middleware.RequireRole("super_admin"))
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	roleSet := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		roleSet[r] = true
	}
	return func(c *gin.Context) {
		roleValue, exists := c.Get(CtxKeyRole)
		if !exists {
			response.FailAuthDenied(c, "无权访问该接口")
			c.Abort()
			return
		}
		role, ok := roleValue.(string)
		if !ok || !roleSet[role] {
			response.FailAuthDenied(c, "无权访问该接口")
			c.Abort()
			return
		}
		c.Next()
	}
}
