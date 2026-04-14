# learn - Go 后端学习与链路复盘

## 作用

这个目录用于沉淀 `server` 的每日学习记录、全链路复盘、缺陷审计与企业级解法。它不是 API 正式契约源，而是帮助我们从“能写代码”推进到“能判断链路是否闭环”的学习型文档。

建议阅读顺序：

1. 先看 [progress.md](./progress.md)，了解当前整体进度。
2. 再看最新链路审计：[lesson-20260414.md](./lesson-20260414.md)。
3. 遇到某个模块不理解时，再回看对应日期的 `lesson-*.md`。

## 与 `server/docs` 的关系

| 文档 | 用途 |
|---|---|
| [../README.md](../README.md) | 后端文档总入口。 |
| [../循序渐进总说明.md](../循序渐进总说明.md) | 12 步落地顺序与每步完成标准。 |
| [../learning-path/go-backend-learning-roadmap.md](../learning-path/go-backend-learning-roadmap.md) | Go 后端学习路线。 |
| [../prd-mapping/README.md](../prd-mapping/README.md) | PRD/技术方案到实现的映射。 |
| [request-chain-frontend-guide.md](./request-chain-frontend-guide.md) | 从前端视角理解后端请求链路。 |

## 当前最重要结论

最新审计日期：2026-04-14。

结论：`server` 的工程骨架链路已经形成，启动、路由、中间件、handler、service、repository、PostgreSQL、MinIO/S3 都有基本闭环；但从企业交付标准看，目前还不能算完整可交付。

关键原因：

- 当前 `go test ./...` 编译失败，阻断点在 `internal/transport/http/helper/utils.go:53`。
- 参数模板 `PUT /admin/param-templates/:id` 存在 URL id 与 body id 契约断点。
- 参数模板 `List/Delete/GetById/GetItemsById` 的 server 链路已补充，但分页空结果、错误处理、路由命名仍需收口。
- 公开端只注册了 `GET /api/v1/public/vehicles`，首页、分类、详情、联系配置、分享兜底尚未闭环。
- 发布校验、媒体补偿/GC、RBAC 挂载、初始管理员种子、CI 门禁仍需补齐。

详细分析见 [lesson-20260414.md](./lesson-20260414.md)。

## 文件索引

| 文件 | 说明 |
|---|---|
| [progress.md](./progress.md) | 总进度、当前全链路状态、每日倒序记录。 |
| [lesson-20260414.md](./lesson-20260414.md) | 全链路关系审计、当前缺陷、参数模板 server 链路补齐复盘、企业级解决方案、验收清单。 |
| [lesson-20260413.md](./lesson-20260413.md) | 参数模板全链路：迁移、模型、GORM 一对多、事务 Update、组合根接入。 |
| [lesson-20260412.md](./lesson-20260412.md) | 依赖注入重构：组合根、`buildHandlers()`、repo/service/handler 装配。 |
| [lesson-20260411.md](./lesson-20260411.md) | `system_settings` 单行表、系统设置读写链路。 |
| [lesson-20260410.md](./lesson-20260410.md) | JWT 认证闭环：双 Token、httpOnly Cookie、Refresh、Logout、RBAC 能力。 |
| [lesson-20260409.md](./lesson-20260409.md) | `media_assets`、上传落库、车辆写接口、封面图 URL 拼接。 |
| [lesson-20260408.md](./lesson-20260408.md) | MinIO/S3 对象存储、上传 Handler、Docker Compose 扩展。 |
| [lesson-20260329.md](./lesson-20260329.md) | struct tag、query/body 绑定、Gin 中间件、参数白名单。 |
| [lesson-20260328.md](./lesson-20260328.md) | GORM 表/列映射、迁移与运行时职责分离。 |
| [lesson-20260325.md](./lesson-20260325.md) | 分类 PATCH/PUT：可选指针体、合并后校验。 |
| [lesson-20260323.md](./lesson-20260323.md) | 301 重定向排障、`*string` 指针修复、JSON 命名风格统一。 |
| [lesson-20260321.md](./lesson-20260321.md) | 数据库迁移、Repository、车辆 List 链路诊断。 |
| [lesson-20260320.md](./lesson-20260320.md) | domain、PostgreSQL 连接、bootstrap 注入。 |
| [lesson-20260319.md](./lesson-20260319.md) | HTTP 基座、router、handler、统一响应。 |
| [lesson-20260318.md](./lesson-20260318.md) | 指针/值与 logger 数据流。 |
| [lesson-20260317.md](./lesson-20260317.md) | 工程壳、配置、日志、响应、启动入口。 |

## 如何判断一条后端链路完整

一条企业级链路不要只看“有没有路由”或“有没有 service”，而要按业务动作逐项检查：

```text
路由
  -> 鉴权/授权
  -> 请求 DTO
  -> 参数校验
  -> 业务规则
  -> repo/事务
  -> DB/OSS/外部依赖
  -> 响应 DTO
  -> 错误码
  -> 日志/审计
  -> 测试
  -> 文档契约
```

如果其中任何一环靠隐式约定、路径字符串判断、前端补救、人工记忆或没有测试兜底，就只能算“能跑”，不能算“企业级闭环”。
