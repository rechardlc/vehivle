package oss

import (
	"fmt"

	"github.com/minio/minio-go/v7"
)

// MinioClient 对象存储（MinIO/S3）客户端及访问配置
type MinioClient struct {
	Endpoint  string
	PublicURL string
	Bucket    string
	Client    *minio.Client
}

// ObjectPublicURL 拼接桶内对象的公网访问地址（与 upload 成功响应中的 url 规则一致）。
func (c MinioClient) ObjectPublicURL(storageKey string) string {
	if c.PublicURL == "" || c.Bucket == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s", c.PublicURL, c.Bucket, storageKey)
}
