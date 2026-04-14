# learn - Go 后端学习与链路复盘

## 作用

这个目录用于沉淀 `server` 的每日学习记录、全链路复盘、缺陷审计与企业级解法。它不是 API 正式契约源，而是帮助我们从“能写代码”推进到“能判断链路是否闭环”的学习型文档。

建议阅读顺序：

1. 先看 [progress.md](./progress.md)，了解当前整体进度。
2. 再看最新未提交改动复盘：[lesson-20260414-uncommitted-review.md](./lesson-20260414-uncommitted-review.md)。
3. 需要理解上一轮全链路审计时，再看 [lesson-20260414.md](./lesson-20260414.md)。
4. 遇到某个模块不理解时，再回看对应日期的 `lesson-*.md`。

## 与 `server/docs` 的关系

| 文档 | 用途 |
|---|---|
| [../README.md](../README.md) | 后端文档总入口。 |
| [../循序渐进总说明.md](../循序渐进总说明.md) | 12 步落地顺序与每步完成标准。 |
| [../learning-path/go-backend-learning-roadmap.md](../learning-path/go-backend-learning-roadmap.md) | Go 后端学习路线。 |
| [../prd-mapping/README.md](../prd-mapping/README.md) | PRD/技术方案到实现的映射。 |
| [request-chain-frontend-guide.md](./request-chain-frontend-guide.md) | 从前端视角理解后端请求链路。 |

## 当前最重要结论

最新复盘日期：2026-04-14。

结论：`server` 的工程骨架链路已经形成，本轮未提交内容继续补齐了公开端读取、车型详情图、参数模板 Update 契约、分类分页默认值和发布校验等关键断点；但从企业交付标准看，目前还不能算完整可交付。

关键原因：

- `helper.RequiredField` 的已知编译阻断点已修复，但仍需要用 `go test ./...` 验证全仓编译与测试。
- 参数模板 `PUT /admin/param-templates/:id` 已改为以 URL id 为权威 id，列表空结果和 `ItemsCount` 错误传播也已收口；`getItemsById/getItemsbyId` 双路由仍需后续统一。
- 公开端已补 `home/categories/vehicles/detail/contact/share-check`，但首页 `banners/zones` 仍是占位，公开 contact 的错误处理也偏静默。
- 车型详情图链路已从迁移、模型、repo/service、管理端接口到后台多图上传形成闭环；但车型保存与详情图保存是两次请求，存在部分成功风险。
- 发布校验已真实检查封面图、启用分类、详情图和必填参数；但 `vehicle_param_values` 仍缺管理端参数值写入/回显闭环。
- 媒体补偿/GC、RBAC 挂载、初始管理员种子、CI 门禁和自动化测试仍需补齐。

详细分析见 [lesson-20260414-uncommitted-review.md](./lesson-20260414-uncommitted-review.md) 与 [lesson-20260414.md](./lesson-20260414.md)。

## 文件索引

| 文件 | 说明 |
|---|---|
| [progress.md](./progress.md) | 总进度、当前全链路状态、每日倒序记录。 |
| [lesson-20260414-uncommitted-review.md](./lesson-20260414-uncommitted-review.md) | 基于未提交 diff 的二次复盘：公开端、车型详情图、发布校验、参数模板契约与剩余风险。 |
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
