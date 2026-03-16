# migrations

## 作用
数据库版本迁移目录。

## 目标
保证 schema 变更可追踪、可回滚、可在多环境重复执行。

## V1 表结构目标
- admin_users
- categories
- param_templates
- param_template_items
- vehicles
- vehicle_param_values
- media_assets
- audit_logs
- system_settings

## 规范
每次迁移都附带变更说明和回滚策略。
