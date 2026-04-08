package handler

import (
	"fmt"
	"path/filepath"
	"time"

	"vehivle/internal/infrastructure/oss"
	"vehivle/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

const maxImageUploadSize int64 = 10 << 20 // 10MB

// allowedImageTypes MIME 白名单，防止上传非图片文件
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// Upload 图片上传处理器
type Upload struct {
	OSS oss.MinioClient
}

// NewUpload 创建图片上传处理器
func NewUpload(ossClient oss.MinioClient) *Upload {
	return &Upload{OSS: ossClient}
}

// UploadImages 处理单张图片上传，校验类型与大小后存入 OSS，返回公开访问 URL 与对象键。
func (u *Upload) UploadImages(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.FailMedia(c, "文件解析失败，请检查上传参数")
		return
	}
	defer file.Close()

	if header.Size > maxImageUploadSize {
		response.FailMedia(c, fmt.Sprintf("图片大小不能超过 %dMB", maxImageUploadSize>>20))
		return
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		response.FailMedia(c, "仅支持 JPEG、PNG、GIF、WebP 格式的图片")
		return
	}

	// 按日期分目录 + UUID 保证唯一性，避免文件名碰撞与特殊字符问题
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	objectName := fmt.Sprintf("images/%s/%s%s",
		time.Now().Format("20060102"),
		uuid.New().String(),
		ext,
	)

	_, err = u.OSS.Client.PutObject(c.Request.Context(), u.OSS.Bucket, objectName, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		response.FailMedia(c, "图片上传失败，请稍后重试")
		return
	}

	response.Success(c, gin.H{
		"url":        fmt.Sprintf("%s/%s/%s", u.OSS.PublicURL, u.OSS.Bucket, objectName),
		"objectName": objectName,
	})
}
