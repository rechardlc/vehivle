package handler

import (
	"github.com/gin-gonic/gin"
	"vehivle/pkg/response"
)

// 获取车型列表
func VehiclesListHandler(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "ok",
	})
}

// 创建车型
func VehiclesCreateHandler(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "ok",
	})
}

// 更新车型
func VehiclesUpdateHandler(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "ok",
	})
}

// 删除车型
func VehiclesDeleteHandler(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "ok",
	})
}