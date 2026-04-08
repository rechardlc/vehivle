# learn — Go 后端学习记录

## 作用
沉淀每日学习笔记、进度与复盘，与 `server/docs` 其他目录配合使用。

## 与 `server/docs` 的关系
| 目录 | 内容 |
|------|------|
| [../README.md](../README.md) | 文档总览与原则 |
| [../循序渐进总说明.md](../循序渐进总说明.md) | 12 步落地顺序与每步完成标准 |
| [../learning-path/go-backend-learning-roadmap.md](../learning-path/go-backend-learning-roadmap.md) | 阶段化学习路线与验收 |
| [../prd-mapping/README.md](../prd-mapping/README.md) | PRD/技术到实现的映射 |

本目录专注 **执行记录**：今天学了什么、代码推进到哪、明天第一件事。

## 文件索引
| 文件 | 说明 |
|------|------|
| [progress.md](./progress.md) | 总进度表、全链路状态、每日倒序记录 |
| [lesson-20260317.md](./lesson-20260317.md) | 知识库（工程壳、配置、日志等） |
| [lesson-20260318.md](./lesson-20260318.md) | 指针/值与 logger 数据流 |
| [lesson-20260319.md](./lesson-20260319.md) | HTTP 基座、router、通配响应 |
| [lesson-20260320.md](./lesson-20260320.md) | domain、Postgres 连接、bootstrap 注入 |
| [lesson-20260321.md](./lesson-20260321.md) | 迁移、分层诊断、车型 List 链路落地 |
| [lesson-20260323.md](./lesson-20260323.md) | 301 重定向排障、air 静默降级、`*string` 指针修复 |
| [lesson-20260325.md](./lesson-20260325.md) | 分类 PATCH：可选指针体、合并与校验、`Name` vs `ParentID` 赋值 |
| [lesson-20260328.md](./lesson-20260328.md) | GORM 表/列映射、`TableName`、迁移与运行时、domain 多类型同文件 |
| [lesson-20260329.md](./lesson-20260329.md) | struct tag（`json`/`form`/`gorm`）、`ShouldBindQuery`/`ShouldBindJSON`、GET query 与 FormData；Gin 中间件（`HandlerFunc`、`Use`、单路由、`Next`/`Abort`、`ValidateParams`） |
| [lesson-20260408.md](./lesson-20260408.md) | MinIO/S3 对象存储集成（客户端封装、Bootstrap 连接池、Bucket 自动创建）；图片上传 Handler（MIME 白名单、`PutObject`、`defer`）；`injectRouter` 错误处理改进；统一响应 `RequestID` 补全；Docker Compose 扩展 MinIO；前端直传适配 |
| [lesson-20260409.md](./lesson-20260409.md) | **`media_assets` 迁移与语义**；上传后写库、返回 `id`/`url`/`storageKey`；`coverImageUrl` 拼接；`VehicleService` 写接口与批量/发布/下架/复制；路由补全；前端 `coverMediaId` = 媒体 UUID |

建议阅读顺序：**先看 [progress.md](./progress.md) 当前状态**，再按日期打开对应 `lesson-*.md`。
