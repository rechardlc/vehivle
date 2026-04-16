# docs

## 作用

这里是 `server` 后端文档入口，用来连接学习记录、落地顺序、PRD 映射和技术方案。

推荐阅读顺序：

1. [learn/progress.md](./learn/progress.md)：当前整体进度、最新提醒和每日倒序记录。
2. [learn/lesson-20260416.md](./learn/lesson-20260416.md)：今日学习记录，请求上下文日志链路。
3. [learn/README.md](./learn/README.md)：学习文档索引和阅读路径。
4. [循序渐进总说明.md](./循序渐进总说明.md)：后端 12 步落地顺序与验收标准。
5. [prd-mapping/README.md](./prd-mapping/README.md)：PRD/tech 到实现模块的映射。
6. [learning-path/README.md](./learning-path/README.md)：Go 后端学习路线入口。

## 当前最新学习主题

最新日期：2026-04-16。

今日主题是请求上下文日志链路：

```text
RequestID
  -> context.Context
  -> slog.Handler
  -> AccessLog
  -> JWT user_id
```

当前重点不是新增业务接口，而是让日志链路具备可追踪性：一次请求要能通过 `request_id` 串起来，受保护接口还应能带上 `user_id`，方便后续排查问题和审计操作。

## 目录说明

- `learn/`：每日学习记录、全链路复盘、缺陷审计、进度跟踪。
- `learning-path/`：Go 后端学习路线。
- `prd-mapping/`：PRD、技术方案与后端实现模块之间的映射。

## 维护约定

- 学习记录优先写入 `learn/lesson-YYYYMMDD.md`。
- 每次新增学习记录后，同步更新 `learn/progress.md` 和 `learn/README.md`。
- 只有正式接口契约、架构约定或 PRD 映射变化时，才同步更新 `doc/tech.md` 或 `prd-mapping/`。
