package postgres

import (
	"context"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
	"vehivle/configs"
	"vehivle/pkg/logger"
)

// Open 打开数据库连接
func Open(cfg *configs.DatabaseConfig, log logger.Logger) (*gorm.DB, error) {
	// 检查DSN: 地址、用户名、密码、数据库名
	if cfg.DSN == "" {
		return nil, fmt.Errorf("database DSN is required")
	}
	// 创建GORM实例
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		// Logger: logger.With(logger.Logger).With(logger.Logger),
	})
	if err != nil {
		return nil, err
	}
	// 获取数据库连接
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// 设置连接池
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	// 设置最大连接数
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	// 设置连接最大生命周期
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	}
	log.Info(context.Background(), "database connected", "dsn", "***")
	return db, nil
}

// Ping 检查数据库连接
func Ping(ctx context.Context, db *gorm.DB, skip bool) error {
	// 如果跳过检查，则返回nil
	if skip {
		return nil
	}
	// 获取数据库连接
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	// 检查数据库连接
	return sqlDB.PingContext(ctx)
}

// close 关闭数据库连接
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// isClosed 检查数据库连接是否已关闭
// func isClosed(err error) bool {
// 	return strings.Contains(err.Error(), "sql: database is closed")
// }
