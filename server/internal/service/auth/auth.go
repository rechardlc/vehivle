package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"vehivle/pkg/jwt"
)

// AdminUser 结构体声明
type AdminUser struct {
	ID       string 
	Username string 
	Role     string
	PasswordHash string
}

// AdminUserRepository 接口声明
type AdminUserRepository interface {
	FindByUsername(ctx context.Context, username string) (*AdminUser, error)
	FindByID(ctx context.Context, id string) (*AdminUser, error)
}

// JWTConfig 结构体声明
type JWTConfig struct {
	Secret string
	RefreshSecret string
	ExpireHours int
	RefreshExpireHours int
}
// LoginResult 结构体声明
type LoginResult struct {
	AccessToken string
	RefreshToken string
	ExpiresIn int // Access Token 有效期（秒）
}
// Service 结构体声明
type Service struct {
	repo AdminUserRepository
	cfg JWTConfig
}
func NewService(repo AdminUserRepository, cfg JWTConfig) *Service {
	return &Service{repo: repo, cfg: cfg}
}


// Login 登录
func (s *Service) Login(ctx context.Context, username string, password string) (*LoginResult, error) {
	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("查找用户失败: %w", err)
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}
	// 比较密码哈希
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("密码错误")
	}
	// 生成 access token 和 refresh token
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
		AccessToken: accessToken,
		RefreshToken: refreshToken,
		ExpiresIn: int(accessDuration.Seconds()),
	}, nil
}

func (s *Service) GetCurrentUser(ctx context.Context, userID string) (*AdminUser, error) {
	return s.repo.FindByID(ctx, userID)
}