// Package configs 管理应用配置的加载、校验与环境变量覆盖。
// 支持 YAML 配置文件 + 环境变量，敏感信息通过 VEHIVLE_* 环境变量注入。
package configs

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

const envPrefix = "VEHIVLE"

// Conf 应用总配置，由 YAML + 环境变量合并而成。
type Conf struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Oss      OssConfig      `mapstructure:"oss"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

// AppConfig 应用基础配置。
type AppConfig struct {
	Env          string `mapstructure:"env"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	LogLevel     string `mapstructure:"log_level"`
}

// DatabaseConfig 数据库连接配置。
type DatabaseConfig struct {
	DSN             string `mapstructure:"dsn"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis 连接配置。
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	Timeout  int    `mapstructure:"timeout"`
}

// OssConfig 对象存储配置（媒体直传、签名 URL 等）。
type OssConfig struct {
	AccessKey  string `mapstructure:"access_key"`
	SecretKey  string `mapstructure:"secret_key"`
	Bucket     string `mapstructure:"bucket"`
	Region     string `mapstructure:"region"`
	SignExpire int    `mapstructure:"sign_expire"` // 签名 URL 有效期（秒）
}

// JWTConfig JWT 认证配置。
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

// Load 加载配置：先加载 .env 文件，再按环境读取 YAML，最后用环境变量覆盖。
// 环境由 VEHIVLE_APP_ENV 或 APP_ENV 决定，默认 dev。
// 注意：启动时工作目录需为 server/，以便正确加载 .env 与 .env.{env}。
func Load() (*Conf, error) {
	// 1. 加载 .env 文件，使 VEHIVLE_* 等变量可用（忽略文件不存在）
	// 读取 .env 文件，将变量写入环境变量
	_ = gotenv.Load(".env")
	env := getEnvValue("VEHIVLE_APP_ENV", "APP_ENV", DefaultAppEnv)
	// 读取 .env.{env} 文件，将变量写入环境变量
	_ = gotenv.Load(".env." + env)

	// 2. 初始化 Viper，设置默认值
	v := viper.New()
	setDefaults(v)

	// 3. 按环境选择 YAML 配置文件
	v.SetConfigType("yaml")
	// 配置多文件是为了兼容，找到一次即可
	// 查找到当前目录下的 conf.{env}.yaml 文件
	v.AddConfigPath(".")
	// 查找到当前目录下的 configs 目录下的 conf.{env}.yaml 文件
	v.AddConfigPath("./configs")
	// 查找到当前目录下的 ../configs 目录下的 conf.{env}.yaml 文件
	v.AddConfigPath("../configs")
	// 设置配置文件名
	v.SetConfigName("conf." + env)
	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file conf.%s.yaml: %w", env, err)
	}

	// 4. 环境变量覆盖：VEHIVLE_APP_PORT、VEHIVLE_DATABASE_DSN 等
	// 从.env读取的VEHIVLE_*环境变量会覆盖配置文件中的同名配置项
	// 这里将VEHIVLE_*环境变量转换为VEHIVLE_APP_PORT、VEHIVLE_DATABASE_DSN等配置项
	// 设置环境变量前缀
	v.SetEnvPrefix(envPrefix)
	// 设置环境变量键替换
	// 将.转换为_
	// viper读取yaml时，将app:port会转化为app.port,所以这里将.转换为_
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// 自动加载环境变量
	v.AutomaticEnv()

	// 反序列化配置文件
	// 将配置文件转换为Conf结构体
	var cfg Conf
	// 将配置文件转换为Conf结构体
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// setDefaults 设置各配置项的默认值，值来源于 default.go 中的常量。
func setDefaults(v *viper.Viper) {
	v.SetDefault("app.env", DefaultAppEnv)
	v.SetDefault("app.port", DefaultAppPort)
	v.SetDefault("app.read_timeout", DefaultAppReadTimeout)
	v.SetDefault("app.write_timeout", DefaultAppWriteTimeout)
	v.SetDefault("app.log_level", DefaultAppLogLevel)

	v.SetDefault("database.dsn", "")
	v.SetDefault("database.max_open_conns", DefaultDBMaxOpenConns)
	v.SetDefault("database.max_idle_conns", DefaultDBMaxIdleConns)
	v.SetDefault("database.conn_max_lifetime", DefaultDBConnMaxLifetime)

	v.SetDefault("redis.addr", DefaultRedisAddr)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", DefaultRedisDB)
	v.SetDefault("redis.timeout", DefaultRedisTimeout)

	v.SetDefault("oss.access_key", "")
	v.SetDefault("oss.secret_key", "")
	v.SetDefault("oss.bucket", "")
	v.SetDefault("oss.region", "")
	v.SetDefault("oss.sign_expire", DefaultOSSSignExpire)

	v.SetDefault("jwt.secret", "")
	v.SetDefault("jwt.expire_hours", DefaultJWTExpireHours)
}

// getEnvValue 按优先级读取环境变量，均未设置时返回最后一个参数作为默认值。
// 示例：getEnvValue("VEHIVLE_APP_ENV", "APP_ENV", "dev") 依次尝试，未命中则返回 "dev"。
func getEnvValue(keys ...string) string {
	// 设置默认值
	defaultVal := DefaultAppEnv
	if len(keys) > 0 {
		// 如果设置了默认值，则返回默认值
		defaultVal = keys[len(keys)-1]
	}
	for i := 0; i < len(keys)-1; i++ {
		// 如果环境变量存在，则返回环境变量
		if v := os.Getenv(keys[i]); v != "" {
			// ToLower 将字符串转换为小写
			// TrimSpace 去除字符串两端的空格
			return strings.ToLower(strings.TrimSpace(v))
		}
	}
	return strings.ToLower(strings.TrimSpace(defaultVal))
}

// Validate 校验配置合法性。
func (c *Conf) Validate() error {
	if c.App.Port <= 0 || c.App.Port > 65535 {
		return fmt.Errorf("app.port must be between 1 and 65535, got %d", c.App.Port)
	}
	return nil
}
