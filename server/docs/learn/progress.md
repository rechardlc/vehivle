# Go 后端学习进度

> 每日进度记录与项目落地状态。学习内容见 [lesson-20260317.md](./lesson-20260317.md)（知识库）、[lesson-20260319.md](./lesson-20260319.md)、[lesson-20260320.md](./lesson-20260320.md)、[lesson-20260321.md](./lesson-20260321.md)、[lesson-20260323.md](./lesson-20260323.md)、[lesson-20260325.md](./lesson-20260325.md)、[lesson-20260328.md](./lesson-20260328.md)、[lesson-20260329.md](./lesson-20260329.md)（最新：struct tag 与绑定、Gin 中间件与 `ValidateParams`）。索引见 [learn/README.md](./README.md)。

---

## 一、项目落地顺序（对齐循序渐进总说明）

| 序号 | 内容 | 状态 |
|------|------|------|
| 1 | 第 1 步：工程壳（configs、logger、response、main、bootstrap） | ✅ 已完成 |
| 2 | 第 2 步：HTTP 基座（路由分组、router、handler、NoRoute/NoMethod 通配、完整中间件链） | ✅ 已完成 |
| 3 | 第 3 步：业务语义（domain/model、enum、rule） | ✅ 已完成 |
| 4 | 第 4 步：数据库与迁移 | 进行中（✅ GORM、迁移、`000001`/`000002`；✅ `VehicleRepo`：`GetById`/`Update`/`List`/`Create`；✅ `CategoryRepo`：CRUD 全链路；⏳ 车型 `Update`/`Delete`、分页与参数校验） |
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
| 5 | PostgreSQL + 迁移 + Repository | 进行中（✅ 连接、迁移、车型/分类 List 读写链路；✅ `Update` 支撑 `Publish`；✅ 分类 CRUD 全链路；⏳ 车型 Update/Delete、分页、事务） |
| 6 | 认证与权限（JWT、RBAC） | 待开始 |
| 7 | 核心业务域（分类、车型、媒体等） | 进行中（✅ 分类 CRUD + 状态枚举 + parentName；⏳ 车型完整 CRUD、媒体） |
| 8 | 缓存、性能、异常兜底 | 待开始 |
| 9 | 测试与质量门禁 | 待开始 |
| 10 | 部署与上线 | 待开始 |

---

## 三、当前全链路状态

```
main ✅
  └─ Bootstrap ✅ (cfg → logger → DB → middleware → router)
       └─ Router ✅ (admin/public 分组, health, 404/405)
            ├─ Handler（车型）✅ List/Create 已接通；⏳ Update/Delete 仍为占位
            │    └─ Service ✅ (VehicleRepo 接口 + Publish + List + Create)
            │         └─ Repository ✅ GetById / Update / List / Create（⏳ Delete 待补）
            │              └─ *gorm.DB ✅
            └─ Handler（分类）✅ List/Create 已接通
                 └─ Service ✅ (CategoryRepo 接口 + List/Create/Update/Delete)
                      └─ Repository ✅ CRUD 全链路
                           └─ *gorm.DB ✅
```

### 待办 / 断裂点

| # | 问题 | 文件 |
|---|------|------|
| 1 | 车型 `Update`/`Delete` 未接 Service / DB | `handler/vehicles.go` |
| 2 | `VehicleRepo` 尚无 `Delete`（及分页） | `repository/postgres/vehicle_repo.go` |
| 3 | 分类 `Update`/`Delete` | ✅ 已接通（PATCH 体为指针字段，见 [lesson-20260325](./lesson-20260325.md)） |
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
├── migrations/         # SQL 迁移 ✅（000001 主表、000002 categories 等）
├── internal/bootstrap/ # 依赖装配 ✅
├── internal/infrastructure/postgres/ # Postgres 连接 ✅
├── internal/domain/    # 领域语义 ✅（enum / model / rule）
├── internal/service/vehicle/ # 业务服务 ✅（接口 + Publish 逻辑）
├── internal/repository/postgres/ # 数据仓储 ✅（车型：GetById、Update、List）
├── internal/transport/http/
│   ├── router/         # 路由 ✅
│   └── handler/        # 处理器 ⏳（车型 List ✅；其余方法占位）
├── pkg/
│   ├── logger/         # 日志 ✅
│   └── response/       # 统一响应 ✅
└── ...
```

---

## 五、下一步建议（对齐 [循序渐进总说明](../循序渐进总说明.md) 第 4～5 步）

1. **车型写操作**：为 `VehicleRepo` 增加 `Create`/`Delete`（及需要的 DTO），Handler 中 `Create`/`Update`/`Delete` 调用 Service。
2. **公开 vs 管理端**：`List` 已用路径区分 `onlyPublished`；后续可改为显式参数或两套 handler，避免隐式依赖 URL。
3. **domain / rule**：扩展下架、删除规则，与 `vehicles.status` CHECK 一致。
4. **第 5 步认证**：登录、JWT、RBAC（见 [learning-path 阶段 4](../learning-path/go-backend-learning-roadmap.md)）。
5. **（可选）** 新迁移 `000002_...`（如 `categories`），勿改已执行的 `000001`。
6. **（可选）** 优雅关闭时 `postgres.Close`；GORM SQL 日志接入统一 logger。

---

## 六、每日进度

> 按日期倒序，最新在前。建议每天结束前：自测通过、写 5 行复盘、记录明天第一件事。

### 2026-03-29

- **学习**：Go **struct tag** 是「给具体库看的字符串」；**`json` / `form` / `gorm`** 由不同读者解析，可并列，**不互相替代**。
- **学习**：**`ShouldBindQuery`** 绑定 **URL query**，依赖字段 **`form:"..."`**；**`ShouldBindJSON`** 绑定 **body JSON**，依赖 **`json:"..."`**。GET 列表仅 query 时应用前者，不能用后者接同一套参数。
- **学习**：GET 的「表单感」来自 **query 的 key=value**；前端 **`FormData`** 多为 **POST multipart**，与 **GET + `ShouldBindQuery`** 不是同一用法。
- **学习**：**`gin.HandlerFunc`** 即 **`func(*gin.Context)`**，业务 handler 与中间件**同类型**；**`router.Use(mw)`** 作用于组/引擎上后续路由，**`GET("", mw, handler)`** 仅单路由链式执行；**`c.Next()`** 放行、**`c.Abort()`** 中断。
- **学习**：**`middleware.ValidateParams(allowFields)`** 为工厂函数，返回 **`HandlerFunc`**，对 **query 键**做白名单校验；分类列表路由 **`categories.GET("", ValidateParams(...), List)`**（见 `router/router.go`）。
- **文档**：新增并扩充 [lesson-20260329.md](./lesson-20260329.md)（含第七节中间件），更新 [learn/README.md](./README.md) 索引；[progress.md](./progress.md) 本段与文首「最新」指向同步。
- **明日第一件事**：车型 `Update`/`Delete` 接 Service + Repo，或对照 `categories` handler 梳理绑定与 DTO 分层注释。

### 2026-03-28

- **学习**：**迁移（golang-migrate）** 与 **GORM** 职责分离——迁移负责 DDL/版本；GORM 运行时**不读取** `migrations/` 目录，仅按 struct、`TableName()`、列 tag 生成 SQL。
- **学习**：表名 **`Category` → `categories`** 来自默认 **复数命名策略**；不规则表名需 **`TableName() string`** 与 `CREATE TABLE` 一致。
- **学习**：列名 **`CreatedAt` → `created_at`** 为默认 **snake_case**；可用 **`gorm:"column:..."`** 显式对齐；**`gorm:"-"`** 表示非表列（如 `ParentName`）。
- **学习**：domain struct **不必**包含表的全部列；**同一文件**内 `Category` + `CategoryCreateInput` + `CategoryUpdateBody` + `CategoryListQuery` 在本项目规模下 **内聚合理**。
- **文档**：新增 [lesson-20260328.md](./lesson-20260328.md)，更新 [learn/README.md](./README.md) 索引；[progress.md](./progress.md) 迁移目录说明补充 `000002`。
- **明日第一件事**：车型 `Update`/`Delete` 接 Service + Repo，或为 `Vehicle` 对齐 `TableName`/列 tag 与迁移。

### 2026-03-25

- **完成**：分类 **PUT 部分更新** 契约与实现对齐——`CategoryUpdateBody` 各字段为 **可选指针** + `omitempty`，区分 JSON **未传**与**传 0/空串**（`status`/`sortOrder`/`level` 等）。
- **完成**：**合并后再校验**——`validateResolvedCategory(model.CategoryCreateInput)` 供 Create 与 Update（合并 `existing` 后）共用，避免重复业务判断。
- **学习**：PATCH 合并时 **`Name`（`*string` → `string`）** 需 `*body.Name`，**`ParentID`（`*string` → `*string`）** 直接指针赋值；与领域模型「名称是值、父级是可空外键」一致。
- **学习**：曾误写 **status 合并条件**（仅在「非法值」时写入）会导致合法 0/1 无法更新；正确路径为 **`body.Status != nil` 则写入，再由统一校验兜底**。
- **文档**：新增 [lesson-20260325.md](./lesson-20260325.md)，更新 [learn/README.md](./README.md) 索引；[progress.md](./progress.md) 待办表中分类 Update/Delete 标为已完成。
- **明日第一件事**：车型 `Update`/`Delete` handler 接 Service + Repo，或复用 PATCH 指针模式。

### 2026-03-23

- **排障**：前端 `GET /api/v1/admin/vehicles` 返回 301 重定向问题
- **修复**：`bootstrap.go` 添加 `r.RedirectTrailingSlash = false`，关闭 Gin 尾斜杠自动重定向（前后端分离 API 项目应关闭）
- **修复**：`vehicle_publish.go` — `CategoryID` 从 `string` 改为 `*string` 后的指针空值判断（`v.CategoryID == nil || *v.CategoryID == ""`）
- **修复**：`handler/vehicles.go` — 新增 `strPtr()` 辅助函数，`CategoryID` 赋值改用指针，修复编译错误
- **发现**：`air` 编译失败后静默降级——继续运行旧二进制，且大量启动日志刷走编译错误信息
- **学习**：Gin radix tree 的 `Group` + 子路由尾斜杠行为；`RedirectTrailingSlash` 的历史由来与企业 API 最佳实践
- **学习**：Go `string → *string` 类型变更的连锁影响（赋值、比较、解引用三类用法需全局排查）
- **经验**：改了代码不生效时，先 `go build ./cmd/api` 验证编译是否通过，再查逻辑
- **完成**：前后端 API JSON 全站统一 camelCase（model json tag、response、handler、admin api 层映射全部清理）
- **完成**：分类状态 `TEXT('enabled'/'disabled')` → `SMALLINT(0/1)`（迁移、enum、handler、admin 同步）
- **完成**：分类模块 CRUD —— handler Create（`ShouldBindJSON`）、List（含 parentName 填充）、Service、Repository
- **学习**：Go struct embedding 消除 DTO 与 Entity 字段重复（`Category` 嵌入 `CategoryCreateInput`）
- **学习**：`json` tag 与 GORM 列名独立——改 json tag 不影响 DB 列映射
- **学习**：`gorm:"-"` 标记虚拟字段（如 `ParentName`）不写入数据库
- **经验**：企业项目应在契约里固定 JSON 命名风格，前后端一致；状态枚举用 0/1 还是字符串需在文档里写死
- **明日第一件事**：车型 `Update`/`Delete` handler 接 Service + Repo，或进入认证闭环（第 5 步）

### 2026-03-21

- **完成**：`deploy/docker/docker-compose.yml`（PostgreSQL + Redis，`pull_policy: never`，默认镜像 `postgres:13`、`redis:latest`）；`deploy/docker/README.md`
- **完成**：`cmd/migrate`（golang-migrate；`iofs` + `os.DirFS` 解决 Windows `file://` 迁移源失败）
- **完成**：`migrations/000001_init_schema`（`admin_users`、`vehicles`、`system_settings`）；字段 `--` 注释与 `COMMENT ON TABLE/COLUMN`
- **完成**：`Makefile`：`migrate-up`、`migrate-down`、`migrate-version`
- **完成**：`migrations/README.md`（注释约定、已执行迁移勿改）
- **完成**：全链路诊断（main → bootstrap → router → handler → service → repository → DB），定位并记录原断裂点
- **完成**：`service/vehicle/service.go` — import 路径 `vehivle/...`；`VehicleRepo` 接口定义在 service 包
- **完成（补充）**：`handler/vehicles.go` — `NewVehicleRepo(db)` → `NewService(repo)`；`VehicleRepo` 实现 `Update`、`List`；`Service.List`；`List`  handler 调用 Service，公开路径仅 `published`
- **学习**：Go 接口哲学 — 接口由消费方定义、隐式实现；与 TypeScript/Java `implements` 的差异
- **学习**：Go 四件套模式 — interface → struct → NewXxx → 指针接收者方法
- **复盘**：迁移与连接分离；原断裂点为 Handler 误将 `*gorm.DB` 注入 Service；现已打通 `GET /api/v1/admin/vehicles` 与 `GET /api/v1/public/vehicles` 的列表查询
- **明日第一件事**：车型 `Create`/`Delete` 接 Service + Repo，或进入认证闭环（第 5 步）

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

*最后更新：2026-03-29（struct tag、Gin 绑定与中间件、`ValidateParams`；lesson-20260329 / README / progress 同步）*
