package handler

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"vehivle/internal/domain/model"
	"vehivle/internal/infrastructure/oss"
	"vehivle/internal/repository/postgres"
	"vehivle/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

const maxImageUploadSize int64 = 10 << 20 // 10MB

const mediaAssetTypeImage = "image"

// allowedImageTypes MIME 白名单，防止上传非图片文件
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

type Upload struct {
	OSS       oss.MinioClient
	mediaRepo *postgres.MediaAssetRepo
}

func NewUpload(ossClient oss.MinioClient, mediaRepo *postgres.MediaAssetRepo) *Upload {
	return &Upload{OSS: ossClient, mediaRepo: mediaRepo}
}

// UploadImages 处理单张图片上传，校验类型与大小后存入 OSS，并写入 media_assets。
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
	storageKey := fmt.Sprintf("images/%s/%s%s",
		time.Now().Format("20060102"),
		uuid.New().String(),
		ext,
	)

	_, err = u.OSS.Client.PutObject(c.Request.Context(), u.OSS.Bucket, storageKey, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		response.FailMedia(c, "图片上传失败，请稍后重试")
		return
	}

	mediaID := uuid.New().String()
	row := &model.MediaAsset{
		ID:         mediaID,
		StorageKey: storageKey,
		MimeType:   contentType,
		FileSize:   header.Size,
		AssetType:  mediaAssetTypeImage,
	}
	if err := u.mediaRepo.Create(c.Request.Context(), row); err != nil {
		msg := "媒体元数据保存失败，请稍后重试"
		if strings.Contains(err.Error(), "does not exist") {
			msg = "数据库未创建 media_assets 表：在 server 目录执行 go run ./cmd/migrate -op up 后再试"
		}
		response.FailMedia(c, msg)
		return
	}

	publicURL := u.OSS.ObjectPublicURL(storageKey)
	response.Success(c, gin.H{
		"id":         row.ID,
		"url":        publicURL,
		"storageKey": storageKey,
	})
}
