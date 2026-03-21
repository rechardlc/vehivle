# cmd

## 作用
放置应用启动入口（程序主入口），这里通常只负责组装，不写业务逻辑。

## 目录规范
- 每个可执行目标单独一个子目录。
- 当前 V1 至少会有 `api` 入口，后续可扩展 worker/cron。
- `migrate`：执行 `migrations/` 下 SQL 迁移（PostgreSQL），见 `migrations/README.md` 与 `make migrate-up`。

## 你应该做什么
1. 先定义启动参数（端口、环境、配置路径）。
2. 调用 `internal/bootstrap` 完成依赖注入。
3. 启动 HTTP Server，并处理优雅退出。
