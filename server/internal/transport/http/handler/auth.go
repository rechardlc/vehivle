package handler

import (
	"errors"
	"net/http"
	"vehivle/configs"
	"vehivle/internal/repository/postgres"
	"vehivle/internal/service/auth"
	"vehivle/pkg/jwt"
	"vehivle/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	CookieNameAccessToken  = "access_token"
	CookieNameRefreshToken = "refresh_token"
	CookiePathAPI          = "/api"
	CookiePathAuth         = "/api/v1/admin/auth"
)

type Auth struct {
	authService *auth.Service
	jwtCfg      configs.JWTConfig
}

// NewAuth 组装登录依赖：用户仓储 + JWT 配置。
func NewAuth(db *gorm.DB, jwtCfg configs.JWTConfig) *Auth {
	repo := postgres.NewUserRepo(db)
	svc := auth.NewService(repo, auth.JWTConfig{
		Secret:             jwtCfg.Secret,
		RefreshSecret:      jwtCfg.RefreshSecret,
		ExpireHours:        jwtCfg.ExpireHours,
		RefreshExpireHours: jwtCfg.RefreshExpireHours,
	})
	return &Auth{authService: svc, jwtCfg: jwtCfg}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 用户登录，成功写入双 Token httpOnly Cookie。
func (h *Auth) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailParam(c, "用户名或密码不能为空")
		return
	}

	result, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			response.FailAuth(c, "用户名或密码错误")
			return
		}
		response.FailBusiness(c, "登录服务异常")
		return
	}

	setAccessTokenCookie(c, result.AccessToken, result.ExpiresIn, h.jwtCfg.CookieSecure, h.jwtCfg.CookieDomain)
	setRefreshTokenCookie(c, result.RefreshToken, result.RefreshExpiresIn, h.jwtCfg.CookieSecure, h.jwtCfg.CookieDomain)
	response.Success(c, gin.H{
		"expiresIn": result.ExpiresIn,
	})
}

// Refresh 用 Refresh Token 续签 Access Token（RT 由浏览器自动携带 Cookie）。
func (h *Auth) Refresh(c *gin.Context) {
	rtString, err := c.Cookie(CookieNameRefreshToken)
	if err != nil || rtString == "" {
		response.FailAuth(c, "缺少刷新令牌")
		return
	}

	result, err := h.authService.RefreshToken(c.Request.Context(), rtString)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			response.FailAuth(c, "刷新令牌已过期，请重新登录")
			return
		}
		response.FailAuth(c, "刷新令牌无效，请重新登录")
		return
	}

	setAccessTokenCookie(c, result.AccessToken, result.ExpiresIn, h.jwtCfg.CookieSecure, h.jwtCfg.CookieDomain)
	response.Success(c, gin.H{
		"expiresIn": result.ExpiresIn,
	})
}

// Me 获取当前登录用户信息（需 JWT 中间件前置）。
func (h *Auth) Me(c *gin.Context) {
	userIDValue, exists := c.Get("current_user_id")
	if !exists {
		response.FailAuth(c, "用户未登录")
		return
	}
	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		response.FailAuth(c, "无效用户信息")
		return
	}

	user, err := h.authService.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		response.FailBusiness(c, "获取用户信息失败")
		return
	}
	response.Success(c, user)
}

// Logout 登出，清除 AT + RT Cookie。
func (h *Auth) Logout(c *gin.Context) {
	clearAuthCookies(c, h.jwtCfg.CookieSecure, h.jwtCfg.CookieDomain)
	response.Success(c, nil)
}

// --- Cookie 工具函数 ---

func setAccessTokenCookie(c *gin.Context, token string, maxAge int, secure bool, domain string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     CookieNameAccessToken, // 访问令牌名称
		Value:    token, // 访问令牌值
		Path:     CookiePathAPI, // 访问令牌路径：只有访问/api路径的请求才会携带cookie
		MaxAge:   maxAge, // 访问令牌最大年龄：单位为秒
		Secure:   secure, // 访问令牌是否安全：是否只通过 HTTPS 传输
		HttpOnly: true, // 访问令牌是否只通过 HTTP 传输：是否只通过 HTTP 传输
		Domain:   domain, // 访问令牌域名
		SameSite: http.SameSiteLaxMode, // 访问令牌是否允许跨域
	})
}

func setRefreshTokenCookie(c *gin.Context, token string, maxAge int, secure bool, domain string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     CookieNameRefreshToken,
		Value:    token,
		Path:     CookiePathAuth,
		MaxAge:   maxAge,
		Secure:   secure,
		HttpOnly: true,
		Domain:   domain,
		SameSite: http.SameSiteLaxMode,
	})
}

// clearAuthCookies 清除 AT + RT Cookie，Path 必须与设置时一致。
func clearAuthCookies(c *gin.Context, secure bool, domain string) {
	setAccessTokenCookie(c, "", 0, secure, domain)
	setRefreshTokenCookie(c, "", 0, secure, domain)
}
