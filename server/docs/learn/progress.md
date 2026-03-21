# Go 后端学习进度

> 每日进度记录与项目落地状态。学习内容见 [lesson-20260317.md](./lesson-20260317.md)（知识库）、[lesson-20260319.md](./lesson-20260319.md)、[lesson-20260320.md](./lesson-20260320.md)、[lesson-20260321.md](./lesson-20260321.md)（最新：迁移、Docker Compose、Go 接口哲学、分层链路诊断）。

---

## 一、项目落地顺序（对齐循序渐进总说明）

| 序号 | 内容 | 状态 |
|------|------|------|
| 1 | 第 1 步：工程壳（configs、logger、response、main、bootstrap） | ✅ 已完成 |
| 2 | 第 2 步：HTTP 基座（路由分组、router、handler、NoRoute/NoMethod 通配、完整中间件链） | ✅ 已完成 |
| 3 | 第 3 步：业务语义（domain/model、enum、rule） | ✅ 已完成 |
| 4 | 第 4 步：数据库与迁移 | 进行中（✅ GORM 连接与注入；✅ golang-migrate、`000001` 表结构、Makefile 迁移命令；✅ Service 层接口设计；⚠️ Handler→Service 注入不匹配；⚠️ Repository 实现不完整；⏳ handler 真实读写） |
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
| 4 | domain 建模（enum、model、rule） | ✅ 已完成 |
| 5 | PostgreSQL + 迁移 + Repository | 进行中（✅ 连接、迁移、Service 接口设计；⚠️ Handler→Service 注入需修、Repository 方法待补全；⏳ 打通首条完整读写链路） |
| 6 | 认证与权限（JWT、RBAC） | 待开始 |
| 7 | 核心业务域（分类、车型、媒体等） | 待开始 |
| 8 | 缓存、性能、异常兜底 | 待开始 |
| 9 | 测试与质量门禁 | 待开始 |
| 10 | 部署与上线 | 待开始 |

---

## 三、当前全链路状态

```
main ✅
  └─ Bootstrap ✅ (cfg → logger → DB → middleware → router)
       └─ Router ✅ (admin/public 分组, health, 404/405)
            └─ Handler ⚠️ (构造方式需改：注入 Service 而非 *gorm.DB；方法体需调用 Service)
                 └─ Service ✅ (接口由消费方定义, Publish 逻辑完整)
                      └─ Repository ⚠️ (GetById 已有, 缺 Update / List / Create / Delete)
                           └─ *gorm.DB ✅
```

### 断裂点

| # | 问题 | 文件 |
|---|------|------|
| 1 | `NewVehicles(db)` 传 `*gorm.DB` 给 `NewService`，类型不匹配 | `handler/vehicles.go` |
| 2 | `postgres.VehicleRepo` 只有 `GetById`，缺 `Update` 等方法 | `repository/postgres/vehicle_repo.go` |
| 3 | Handler 方法全为空壳，未调用 Service | `handler/vehicles.go` |
| 4 | User 模块无 Service / Repository 层 | `handler/user.go` |

---

## 四、当前目录结构（带状态）

```
server/
├── go.mod              # 模块定义
├── go.sum              # 依赖校验和
├── Makefile            # 工程命令 ✅（含 migrate-up/down/version）
├── .air.toml           # Air 热更新 ✅
├── cmd/api/main.go     # API 入口 ✅
├── cmd/migrate/        # 数据库迁移 CLI ✅（iofs）
├── configs/            # 配置 ✅
├── deploy/docker/      # 本地 PG/Redis Compose ✅
├── migrations/         # SQL 迁移 ✅（000001 三张主表）
├── internal/bootstrap/ # 依赖装配 ✅
├── internal/infrastructure/postgres/ # Postgres 连接 ✅
├── internal/domain/    # 领域语义 ✅（enum / model / rule）
├── internal/service/vehicle/ # 业务服务 ✅（接口 + Publish 逻辑）
├── internal/repository/postgres/ # 数据仓储 ⚠️（GetById 可用，方法待补全）
├── internal/transport/http/
│   ├── router/         # 路由 ✅
│   └── handler/        # 处理器 ⚠️（构造与方法体待修）
├── pkg/
│   ├── logger/         # 日志 ✅
│   └── response/       # 统一响应 ✅
└── ...
```

---

## 五、下一步建议

1. **补全 Repository**：为 `postgres.VehicleRepo` 添加 `Update`、`List`、`Create`、`Delete` 方法，使其满足 service 定义的接口。
2. **修复 Handler 构造与注入**：`NewVehicles` 接收 `*vehicle.Service`（而非 `*gorm.DB`）；Router 中按 `repo → service → handler` 顺序构建。
3. **打通首条完整链路**：`GET /api/v1/admin/vehicles` → Handler.List → Service.List → Repo.List → DB → 返回真实数据。
4. **domain 延续**：rule 扩展（下架、删除）与表字段一致。
5. **（可选）** 新表迁移 `000002_...`（如 `categories`），勿改已执行的 `000001`。
6. **（可选）** `main` 退出时 `postgres.Close`；GORM SQL 日志接入统一 logger。

---

## 五、每日进度

> 按日期倒序，最新在前。建议每天结束前：自测通过、写 5 行复盘、记录明天第一件事。

### 2026-03-21

- **完成**：`deploy/docker/docker-compose.yml`（PostgreSQL + Redis，`pull_policy: never`，默认镜像 `postgres:13`、`redis:latest`）；`deploy/docker/README.md`
- **完成**：`cmd/migrate`（golang-migrate；`iofs` + `os.DirFS` 解决 Windows `file://` 迁移源失败）
- **完成**：`migrations/000001_init_schema`（`admin_users`、`vehicles`、`system_settings`）；字段 `--` 注释与 `COMMENT ON TABLE/COLUMN`
- **完成**：`Makefile`：`migrate-up`、`migrate-down`、`migrate-version`
- **完成**：`migrations/README.md`（注释约定、已执行迁移勿改）
- **完成**：全链路诊断（main → bootstrap → router → handler → service → repository → DB），定位 4 处断裂点
- **完成**：`service/vehicle/service.go` 修复 — import 路径改正（`vehicle` → `vehivle`）；接口 `VehicleRepo` 定义在消费方（service 包内）
- **学习**：Go 接口哲学 — 接口由消费方定义、隐式实现（鸭子类型）；与 TypeScript/Java `implements` 的差异
- **学习**：Go 四件套模式 — interface → struct → NewXxx → 指针接收者方法；「结构体」（`struct`）vs「结构」（架构）的概念区分
- **复盘**：迁移与连接分离；链路断裂主因是 Handler 直接传 `*gorm.DB` 给 Service（类型不匹配）；接口归属 service 包后 `repository/` 无需单独定义接口文件
- **明日第一件事**：补全 Repository 方法 → 修复 Handler 构造注入 → 打通 `GET /api/v1/admin/vehicles` 首条完整读写链路

### 2026-03-20

- **完成**：domain 建模落地（`internal/domain/enum`：VehicleStatus、PriceMode；`model`：Vehicle 实体及 Publish/Unpublish；`rule`：CanPublishVehicle 发布校验）
- **完成**：`internal/infrastructure/postgres`：`Open`（DSN、连接池、连接日志脱敏）、`Ping`、`Close`
- **完成**：`bootstrap.Run`：`pgsqlConnPool` → `postgres.Ping` → `router.New(r, logger, db).Register()`（启动顺序与 fail fast）
- **完成**：`router`：`New(engine, logger, *gorm.DB)`；`Register` 内 `NewUser`/`NewVehicles` 各建单例并绑定 `Get`/`List`/`Create`/`Update`/`Delete`
- **完成**：`handler/user.go`、`handler/vehicles.go`：`User`/`Vehicles` 结构体持有 `*gorm.DB`，构造函数注入（企业常见写法）
- **完成**：`Makefile` 落地（`deps`/`tidy`/`build`/`run`/`dev`/`test`/`fmt`/`vet`/`check`/`clean`；`VEHIVLE_APP_ENV` 可覆盖；Windows 与类 Unix 分支）
- **复盘**：第 3 步业务语义已启动；第 4 步连接层已接通，handler 仍为占位，待迁移与 Repository；`public/vehicles` 可复用 `vehiclesHandler` 减少重复 `NewVehicles`
- **明日第一件事**：SQL 迁移与表结构，或 domain rule 扩展（下架、删除）

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

*最后更新：2026-03-21（迁移、Docker Compose、Go 接口哲学、分层链路诊断与断裂点修复方向、lesson-20260321 与 progress 同步）*
