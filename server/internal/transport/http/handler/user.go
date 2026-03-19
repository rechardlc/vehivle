package handler
import (
	"vehivle/pkg/response"
	"github.com/gin-gonic/gin"
	"time"
)

// 获取用户信息
func UserHandler(c *gin.Context) {
	userId := c.Param("user_id")
	if userId == "" {
		response.FailParam(c, "user_id is required")
		return
	}
	response.Success(c, gin.H{
		"user_id": userId,
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

// 获取用户列表
func UserListHandler(c *gin.Context) {
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
func UserCreateHandler(c *gin.Context) {
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
func UserUpdateHandler(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "User updated successfully",
	})
}

// 删除用户
func UserDeleteHandler(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "User deleted successfully",
	})
}