# migrations

## 作用
数据库版本迁移目录。

## 目标
保证 schema 变更可追踪、可回滚、可在多环境重复执行。

## V1 表结构目标（完整清单）
- admin_users
- categories
- param_templates
- param_template_items
- vehicles
- vehicle_param_values
- media_assets
- audit_logs
- system_settings

当前迁移顺序简述：
- **`000001_init_schema`**：`admin_users`、`vehicles`、`system_settings`
- **`000002_add_categories_table`**：`categories`，并与 `vehicles.category_id` 外键对齐
- **`000003_create_media_assets_table`**：`media_assets`（上传落库元数据，`storage_key` 唯一）
- **`000004_alter_system_settings`**：重建 `system_settings` 为单行全局配置列
- **`000005_create_param_templates_tables`**：`param_templates`、`param_template_items`（一级分类绑定模板，见 `doc/tech.md` 2.1）

其余规划表在后续迁移中按序追加（**不要**修改已发布环境执行过的旧 `.sql`）。

## 规范
- 每次迁移都附带变更说明和回滚策略。
- 字段注释：`CREATE TABLE` 内用 `--` 行注释说明每列含义；需要被 `\d+`、客户端展示的，可补充 `COMMENT ON TABLE` / `COMMENT ON COLUMN`（见 `000001_init_schema.up.sql`）。
- **已执行过的迁移文件不要改内容**；若库已应用 `000001` 后再补注释，应新增 `00000N_...` 迁移只写 `COMMENT ON ...`，而不是改 `000001`。
- 文件命名：`{版本号}_{简述}.up.sql` / `.down.sql`（版本号递增，例如 `000002_xxx`）。
- 工具：**golang-migrate**（`github.com/golang-migrate/migrate/v4`），由 `cmd/migrate` 调用，使用与 API 相同的 `configs.Load()` 与 `VEHIVLE_DATABASE_DSN`。

## 如何执行（在 `server/` 目录）

1. 确保 PostgreSQL 已启动，`.env` / `.env.dev` 中 `VEHIVLE_DATABASE_DSN` 正确。
2. 拉取依赖（首次或更新 `go.mod` 后）：

   ```bash
   go mod tidy
   ```

3. 应用迁移（up）：

   ```bash
   make migrate-up
   ```

   或：

   ```bash
   go run ./cmd/migrate -op up
   ```

4. 查看当前版本：

   ```bash
   make migrate-version
   ```

5. 回滚一步（开发环境慎用）：

   ```bash
   make migrate-down
   ```

   或指定步数：`go run ./cmd/migrate -op down -steps 1`

## 实现说明
- 迁移元数据表：`schema_migrations`（由 golang-migrate 自动创建）。
- 已应用过 `up` 的迁移文件**不要改内容**；修正问题请新增 `00000N_...` 迁移。
