package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"vehivle/pkg/response"
)

// User 用户处理器
type User struct {
	DB *gorm.DB
}

// NewUser 创建用户处理器
func NewUser(db *gorm.DB) *User {
	return &User{DB: db}
}

// Get 获取用户信息
func (u *User) Get(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "ok",
	})
}

// List 获取用户列表
func (u *User) List(c *gin.Context) {
	response.Success(c, gin.H{
		"users": []gin.H{
			{
				"userId": "1",
				"name":   "John Doe",
			},
		},
	})
}

// Create 创建用户
func (u *User) Create(c *gin.Context) {
	userID := c.Param("user_id")
	response.Success(c, gin.H{
		"userId":    userID,
		"name":      "John Doe",
		"email":     "john.doe@example.com",
		"phone":     "1234567890",
		"address":   "123 Main St, Anytown, USA",
		"city":      "Anytown",
		"state":     "CA",
		"zip":       "12345",
		"country":   "USA",
		"createdAt": time.Now().Format(time.RFC3339),
		"updatedAt": time.Now().Format(time.RFC3339),
	})
}

// Update 更新用户
func (u *User) Update(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "User updated successfully",
	})
}

// Delete 删除用户
func (u *User) Delete(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "User deleted successfully",
	})
}
