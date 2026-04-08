package oss

import "github.com/minio/minio-go/v7"

// MinioClient 对象存储（MinIO/S3）客户端及访问配置
type MinioClient struct {
	Endpoint  string
	PublicURL string
	Bucket    string
	Client    *minio.Client
}
