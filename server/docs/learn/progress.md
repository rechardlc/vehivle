# Go 后端学习进度

> 每日进度记录与项目落地状态。学习内容见 [lesson-20260317.md](./lesson-20260317.md)（知识库）、[lesson-20260319.md](./lesson-20260319.md)、[lesson-20260320.md](./lesson-20260320.md)（最新）。

---

## 一、项目落地顺序（对齐循序渐进总说明）

| 序号 | 内容 | 状态 |
|------|------|------|
| 1 | 第 1 步：工程壳（configs、logger、response、main、bootstrap） | ✅ 已完成 |
| 2 | 第 2 步：HTTP 基座（路由分组、router、handler、NoRoute/NoMethod 通配、完整中间件链） | ✅ 已完成 |
| 3 | 第 3 步：业务语义（domain/model、enum、rule） | 进行中 |
| 4 | 第 4 步：数据库与迁移 | 待开始 |
| 5 | 第 5 步：认证闭环 | 待开始 |
| 6 | 第 6～12 步 | 待开始 |

---

## 二、学习路线阶段

| 阶段 | 目标 | 状态 |
|------|------|------|
| 0 | 环境、术语、项目边界 | ✅ 已完成 |
| 1 | Go 基础 + 最小可运行 + 健康检查 | ✅ 已完成 |
| 2 | configs、logger、response、main 启动、中间件 | ✅ 已完成 |
| 3 | bootstrap 抽取、路由分组（admin/public）、完整中间件链 | ✅ 已完成 |
| 4 | domain 建模（enum、model、rule） | 进行中 |
| 5 | PostgreSQL + 迁移 + Repository | 待开始 |
| 6 | 认证与权限（JWT、RBAC） | 待开始 |
| 7 | 核心业务域（分类、车型、媒体等） | 待开始 |
| 8 | 缓存、性能、异常兜底 | 待开始 |
| 9 | 测试与质量门禁 | 待开始 |
| 10 | 部署与上线 | 待开始 |

---

## 三、当前目录结构（带状态）

```
server/
├── go.mod              # 模块定义
├── go.sum              # 依赖校验和
├── .air.toml           # Air 热更新配置 ✅（cmd 指向 ./cmd/api）
├── cmd/api/main.go     # 入口 ✅（configs.Load、bootstrap.Run、端口从配置读取）
├── configs/            # 配置 ✅（YAML + env、viper、Conf 结构、校验）
├── internal/bootstrap/ # 依赖装配 ✅（Gin 初始化、中间件链、router 注册）
├── internal/domain/    # 领域语义 ✅（enum、model、rule 已落地）
│   ├── enum/           # 枚举 ✅（VehicleStatus、PriceMode）
│   ├── model/          # 实体 ✅（Vehicle、Publish/Unpublish 方法）
│   └── rule/           # 业务规则 ✅（CanPublishVehicle）
├── internal/transport/http/
│   ├── router/         # 路由 ✅（admin/user、admin/vehicles、public/vehicles）
│   └── handler/        # 处理器 ✅（user.go、vehicles.go 已落地）
├── pkg/
│   ├── logger/         # 日志 ✅（Logger、RequestID、AccessLog 中间件）
│   └── response/       # 统一响应 ✅
└── ...
```

---

## 四、下一步建议（第 3 步：业务语义）

1. **domain 补齐**：补充 enum 合法性校验函数、扩展 rule（如下架规则、删除规则）
2. **handler 接入 domain**：vehicles handler 调用 `rule.CanPublishVehicle` 等，为后续 service 打基础
3. **（可选）自定义 Recovery**：用 `pkg/response` 统一格式返回 panic 时的 500 响应
4. **（可选）中间件抽取**：若中间件增多，可抽到 `internal/transport/http/middleware/`

---

## 五、每日进度

> 按日期倒序，最新在前。建议每天结束前：自测通过、写 5 行复盘、记录明天第一件事。

### 2026-03-20

- **完成**：domain 建模落地（`internal/domain/enum`：VehicleStatus、PriceMode；`model`：Vehicle 实体及 Publish/Unpublish；`rule`：CanPublishVehicle 发布校验）
- **完成**：vehicles handler 抽取（`handler/vehicles.go`：VehiclesListHandler、Create、Update、Delete）
- **完成**：admin/vehicles 路由挂载（GET/POST/PUT/DELETE），public/vehicles 复用 VehiclesListHandler
- **复盘**：第 3 步业务语义已启动，enum/model/rule 与 PRD 状态机对齐；handler 占位就绪，待接入 service
- **明日第一件事**：domain rule 扩展（下架、删除规则），或进入第 4 步数据库与迁移

### 2026-03-19

- **完成**：Air 热更新配置（`.air.toml` 构建命令修正为 `./cmd/api`，开发可用 `air` 替代 `go run`）
- **完成**：bootstrap 落地（Gin 初始化、中间件链、端口从配置读取）
- **完成**：main 精简为 configs.Load → bootstrap.New → b.Run → r.Run
- **完成**：router 包落地（Router 结构体、New、Register、healthHandler 移入 router）
- **完成**：路由分组 `/api/v1/admin`、`/api/v1/public`，健康检查 `GET /health`
- **完成**：handler 抽取（`handler/user.go`：UserHandler 等，admin/user 已挂载）
- **完成**：NoRoute/NoMethod 通配（未注册路由返回统一 JSON 404，方法不匹配返回 405）
- **完成**：response 扩展（CodeNotFound、CodeMethodNotAllowed、FailNotFound、FailMethodNotAllowed）
- **复盘**：第 2 步 HTTP 基座核心完成，handler 与 router 职责清晰，通配保证未注册请求统一响应；Air 热更新提升开发效率
- **明日第一件事**：public/vehicles handler 抽取，或进入第 3 步 domain 建模

### 2026-03-17

- **完成**：第 1 步工程壳落地（configs、logger、response、main 完整启动）
- **完成**：configs 配置加载（YAML + env、viper、Conf 结构）
- **完成**：logger 中间件（RequestID、AccessLog）、logger.Env 类型转换
- **完成**：lesson-20260317 更新（configs、中间件、Env 类型、main 流程）
- **复盘**：第 1 步完成，服务可启动、健康检查可访问、日志与响应结构稳定
- **明日第一件事**：第 2 步 HTTP 基座（路由分组 admin/public）

### 2026-03-18（历史）

- **完成**：Go 指针/值/类型概念梳理（前端视角）
- **完成**：lesson-20260317 更新（补充 5.4 指针深入、7.4 logger 内部实现）
- **复盘**：理解 `*T` 与 `T`、`&` 取地址、接收者是指针时 `l` 已是指针、With 方法数据流

---

*最后更新：2026-03-20（domain 建模、vehicles handler、admin/vehicles 路由）*
