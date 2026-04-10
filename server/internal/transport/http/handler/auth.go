package handler

import (
	"net/http"
	"vehivle/internal/service/auth"
	"vehivle/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


type Auth struct {
	authService *auth.Service
}

func NewAuth(db *gorm.DB) *Auth {
	authService := auth.NewAuth(db)
	return &Auth{authService: authService}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Auth) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailParam(c, "用户名或密码不能为空")
		return
	}
	result, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.Fail(c, response.CodeBusinessError, "用户名或密码错误")
		return
	}
	setAccessTokenCookie(c, result.AccessToken, result.ExpiresIn, true, "")
	setRefreshTokenCookie(c, result.RefreshToken, result.ExpiresIn, true, "")
	response.Success(c, gin.H{
		"expiresIn": result.ExpiresIn,
	})
}

// func (h *Auth) Refresh(c *gin.Context) {
// 	userIDValue, exists := c.Get("current_user_id")
// 	if !exists {
// 		response.FailAuth(c, "用户未登录")
// 		return
// 	}
// 	userID, ok := userIDValue.(string)
// 	if !ok || userID == "" {
// 		response.FailAuth(c, "无效用户信息")
// 		return
// 	}
// 	result, err := h.authService.RefreshToken(c.Request.Context(), userID)
// 	if err != nil {
// 		response.Fail(c, response.CodeBusinessError, "刷新 token 失败")
// 		return
// 	}
// 	setAccessTokenCookie(c, result.AccessToken, result.ExpiresIn, true, "")
// 	response.Success(c, gin.H{
// 		"expiresIn": result.ExpiresIn,
// 	})
// }
func setAccessTokenCookie(c *gin.Context, token string, maxAge int, secure bool, domain string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name: "access_token",
		Value: token,
		Path: "/api",
		MaxAge: maxAge,
		Secure: secure,
		HttpOnly: true,
		Domain: domain,
		SameSite: http.SameSiteNoneMode,
	})
}

func setRefreshTokenCookie(c *gin.Context, token string, maxAge int, secure bool, domain string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name: "refresh_token",
		Value: token,
		Path: "/api/v1/admin/auth",
		MaxAge: maxAge,
		Secure: secure,
		HttpOnly: true,
		Domain: domain,
		SameSite: http.SameSiteNoneMode,
	})
}
func ClearAccessTokenCookie(c *gin.Context) {
	setAccessTokenCookie(c, "", 0, true, "")
	setRefreshTokenCookie(c, "", 0, true, "")
}