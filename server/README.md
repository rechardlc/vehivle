# Server 目录总览

## 目标
这个目录是两/三轮车渠道分销电子展厅 V1 的 Go 后端工作区。
你会在这里手写所有后端代码，本目录先提供企业项目常见骨架和职责说明，不预生成任何业务代码。

## 架构原则
- 分层遵循需求文档的 Handler -> Service -> Repository -> Domain。
- 接口分为后台管理 API 与小程序公开 API 两条线。
- 先保证可维护和可测试，再追求功能堆叠。
- 与 PRD 对齐：认证、车型、分类、参数模板、媒体、系统设置、审计、公开查询。

## 建议你按这个顺序落地
1. `cmd/api` + `internal/bootstrap`：先把服务能跑起来。
2. `configs` + `pkg/logger` + `pkg/response`：把工程基础设施稳定住。
3. `internal/domain` + `internal/repository`：先建数据模型和数据访问边界。
4. `internal/service`：逐步实现业务规则。
5. `internal/transport/http`：按 API 文档接入路由、处理器、中间件。
6. `tests` + `api/openapi`：同步补测试与接口契约。
7. `deploy`：最后处理容器化和上线策略。

## 和 PRD/tech 的关系
- 详细映射见 `docs/prd-mapping/README.md`。
- Go 学习与实战路径见 `docs/learning-path/go-backend-learning-roadmap.md`。
- 学习记录与每日进度见 `docs/learn/progress.md`。
