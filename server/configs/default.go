// Package configs 的默认值常量，供 setDefaults 与业务代码引用。
// 新增配置项时在此补充，并同步更新 setDefaults。
package configs

// App 默认值
const (
	DefaultAppEnv          = "dev"
	DefaultAppPort         = 9999
	DefaultAppReadTimeout  = 30
	DefaultAppWriteTimeout = 30
	DefaultAppLogLevel     = "debug"
)

// Database 默认值
const (
	DefaultDBMaxOpenConns    = 25
	DefaultDBMaxIdleConns    = 5
	DefaultDBConnMaxLifetime = 300
)

// Redis 默认值
const (
	DefaultRedisAddr    = "localhost:6379"
	DefaultRedisDB      = 0
	DefaultRedisTimeout = 3
)

// OSS 默认值
const (
	DefaultOSSSignExpire       = 3600
	DefaultOSSRegion           = "us-east-1"
	DefaultOSSBucket           = "vehivle-media"
	DefaultOSSEnablePublicRead = false
)

// JWT 默认值
const (
	DefaultJWTExpireHours = 24
)
