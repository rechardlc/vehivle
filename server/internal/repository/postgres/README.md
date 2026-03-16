# internal/repository/postgres

## 作用
实现 PostgreSQL 读写与查询组合。

## 对应核心表
- admin_users
- categories
- param_templates / param_template_items
- vehicles / vehicle_param_values
- media_assets
- audit_logs
- system_settings

## 重点实践
- 索引设计对齐 tech 文档。
- 分页查询统一封装，避免重复 SQL/GORM 逻辑。
- 软删除（逻辑删除）策略要统一。
