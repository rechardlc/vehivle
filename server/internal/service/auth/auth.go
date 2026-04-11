package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"vehivle/internal/domain/model"

	"vehivle/pkg/jwt"
)

var ErrInvalidCredentials = errors.New("用户名或密码错误")

// AdminUserRepository 后台用户仓储（登录、当前用户查询）。
type AdminUserRepository interface {
	FindByUsername(ctx context.Context, username string) (*model.AdminUser, error)
	FindByID(ctx context.Context, id string) (*model.AdminUser, error)
}

// JWTConfig 认证所需的 JWT 配置。
type JWTConfig struct {
	Secret             string
	RefreshSecret      string
	ExpireHours        int
	RefreshExpireHours int
}

// LoginResult 登录成功后的 Token 与有效期信息。
type LoginResult struct {
	AccessToken      string
	RefreshToken     string
	ExpiresIn        int // Access Token 有效期（秒）
	RefreshExpiresIn int // Refresh Token 有效期（秒）
}

// Service 认证业务服务。
type Service struct {
	repo AdminUserRepository
	cfg  JWTConfig
}

func NewService(repo AdminUserRepository, cfg JWTConfig) *Service {
	return &Service{repo: repo, cfg: cfg}
}

// Login 用户名+密码登录，成功签发双 Token。
// 失败统一返回 ErrInvalidCredentials，不泄露具体原因（用户不存在 / 密码错误）。
func (s *Service) Login(ctx context.Context, username string, password string) (*LoginResult, error) {
	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("查找用户失败: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessDuration := time.Duration(s.cfg.ExpireHours) * time.Hour
	refreshDuration := time.Duration(s.cfg.RefreshExpireHours) * time.Hour

	accessToken, err := jwt.GenerateToken(s.cfg.Secret, accessDuration, user.ID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("生成 access token 失败: %w", err)
	}
	refreshToken, err := jwt.GenerateToken(s.cfg.RefreshSecret, refreshDuration, user.ID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("生成 refresh token 失败: %w", err)
	}

	return &LoginResult{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        int(accessDuration.Seconds()),
		RefreshExpiresIn: int(refreshDuration.Seconds()),
	}, nil
}

// RefreshToken 用 Refresh Token 签发新的 Access Token。
// RT 过期或无效则返回错误，调用方应引导用户重新登录。
func (s *Service) RefreshToken(ctx context.Context, refreshTokenString string) (*LoginResult, error) {
	claims, err := jwt.ParseToken(s.cfg.RefreshSecret, refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("refresh token 无效: %w", err)
	}

	user, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	accessDuration := time.Duration(s.cfg.ExpireHours) * time.Hour
	accessToken, err := jwt.GenerateToken(s.cfg.Secret, accessDuration, user.ID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("生成 access token 失败: %w", err)
	}

	return &LoginResult{
		AccessToken: accessToken,
		ExpiresIn:   int(accessDuration.Seconds()),
	}, nil
}

// GetCurrentUser 根据已验证的 userID 查询用户信息。
func (s *Service) GetCurrentUser(ctx context.Context, userID string) (*model.AdminUser, error) {
	return s.repo.FindByID(ctx, userID)
}
