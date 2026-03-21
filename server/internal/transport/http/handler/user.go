package handler
import (
	"vehivle/pkg/response"
	"github.com/gin-gonic/gin"
	"time"
	"gorm.io/gorm"
)
// User 用户处理器
type User struct {
	DB *gorm.DB
}

// NewUser 创建用户处理器
func NewUser(db *gorm.DB) *User {
	return &User{DB: db}
}
// 获取用户信息
func (u *User) Get(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "ok",
	})
}

// 获取用户列表
func (u *User) List(c *gin.Context) {
	response.Success(c, gin.H{
		"users": []gin.H{
			{
				"user_id": "1",
				"name": "John Doe",
			},
		},
	})
}

// 创建用户
func (u *User) Create(c *gin.Context) {
	userID := c.Param("user_id")
	response.Success(c, gin.H{
		"user_id": userID,
		"name": "John Doe",
		"email": "john.doe@example.com",
		"phone": "1234567890",
		"address": "123 Main St, Anytown, USA",
		"city": "Anytown",
		"state": "CA",
		"zip": "12345",
		"country": "USA",
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
	})
}

// 更新用户
func (u *User) Update(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "User updated successfully",
	})
}

// 删除用户
func (u *User) Delete(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "User deleted successfully",
	})
}