# Go 后端学习进度

> 每日进度记录与项目落地状态。学习内容见 [lesson-20260317.md](./lesson-20260317.md)（知识库）、[lesson-20260319.md](./lesson-20260319.md)、[lesson-20260320.md](./lesson-20260320.md)、[lesson-20260321.md](./lesson-20260321.md)、[lesson-20260323.md](./lesson-20260323.md)、[lesson-20260325.md](./lesson-20260325.md)、[lesson-20260328.md](./lesson-20260328.md)、[lesson-20260329.md](./lesson-20260329.md)、[lesson-20260408.md](./lesson-20260408.md)、[lesson-20260409.md](./lesson-20260409.md)、[lesson-20260410.md](./lesson-20260410.md)、[lesson-20260411.md](./lesson-20260411.md)、[lesson-20260412.md](./lesson-20260412.md)、[lesson-20260413.md](./lesson-20260413.md)（最新：**参数模板全链路**——`000005` 迁移、`ParamTemplate` + `ParamTemplateItem` 建模、GORM 一对多事务更新、`buildHandlers()` 新模块接入）。索引见 [learn/README.md](./README.md)。

---

## 一、项目落地顺序（对齐循序渐进总说明）

| 序号 | 内容 | 状态 |
|------|------|------|
| 1 | 第 1 步：工程壳（configs、logger、response、main、bootstrap） | ✅ 已完成 |
| 2 | 第 2 步：HTTP 基座（路由分组、router、handler、NoRoute/NoMethod 通配、完整中间件链） | ✅ 已完成 |
| 3 | 第 3 步：业务语义（domain/model、enum、rule） | ✅ 已完成 |
| 4 | 第 4 步：数据库与迁移 | 进行中（✅ GORM、迁移 **`000001`～`000005`**（含 media_assets、system_settings 单行表、**param_templates + param_template_items**）；✅ `VehicleRepo`/`CategoryRepo`/`SystemRepo`/`MediaAssetRepo`/`ParamTemplateRepo`；✅ 车型 Update/Delete（逻辑删除）；⏳ 分页与参数校验） |
| 4.5 | 第 4.5 步：对象存储（媒体上传） | ✅ 已完成（MinIO 客户端封装、Bootstrap 连接池、Bucket 自动创建、图片上传 Handler、前端直传适配） |
| 5 | 第 5 步：认证闭环 | ✅ 已完成（JWT 双 Token + httpOnly Cookie + JWTAuth 中间件 + RequireRole 角色鉴权 + Refresh 续签 + Logout 清除 + 安全加固） |
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
| 5 | PostgreSQL + 迁移 + Repository | 进行中（✅ 连接、迁移至 **`000005`**、车型/分类/系统设置/参数模板读写链路；✅ `Update` 支撑 `Publish`；✅ 分类 CRUD；✅ 车型 Update/逻辑 Delete；✅ MediaAssetRepo；✅ SystemRepo；✅ **ParamTemplateRepo**（含事务 Update）；⏳ 分页） |
| 5.5 | 对象存储（MinIO/S3） | ✅ 已完成（客户端封装、连接池、Bucket 管理、图片上传 Handler、Docker MinIO 服务） |
| 6 | 认证与权限（JWT、RBAC） | ✅ 已完成（双 Token + httpOnly Cookie + JWTAuth + RequireRole + Refresh + Logout + Validate 加固） |
| 7 | 核心业务域（分类、车型、媒体、系统设置、参数模板等） | 进行中（✅ 分类 CRUD；✅ 上传 MinIO + media_assets 落库；✅ 车型 List/Create/Update/Delete/Publish/Unpublish/Duplicate/Batch；✅ 系统设置 Detail/Create/Update + 管理端页面；✅ **参数模板 Create/Update/Detail/AllDetail** + 管理端页面（一级分类绑定）；⏳ 参数模板 List/Delete 路由；⏳ 公开端联系配置/分享兜底） |
| 8 | 缓存、性能、异常兜底 | 待开始 |
| 9 | 测试与质量门禁 | 待开始 |
| 10 | 部署与上线 | 待开始 |

---

## 三、当前全链路状态

```
main ✅
  └─ Bootstrap ✅ (cfg → Validate → logger → DB → Ping → OSS → buildHandlers → router)
       │
       ├─ buildHandlers() ✅  ← 组合根：集中装配 repo → service → handler
       │    ├─ Repos:     NewUserRepo / NewCategoryRepo / NewVehicleRepo / NewSysSettings / NewMediaAssetRepo / NewParamTemplateRepo
       │    ├─ Services:  auth.NewService / category.NewCategoryService / vehicle.NewService / system_setting.NewSysService / param_template.NewParamTemplateService
       │    └─ Handlers:  NewAuth(svc) / NewUser() / NewVehicles(svc,svc,repo,oss) / NewCategories(svc) / NewSysSettings(svc,oss) / NewUpload(oss,repo) / NewParamTemplates(svc)
       │
       └─ Router ✅ (收 Handlers 结构体，不碰 *gorm.DB)
            ├─ Auth 路由组（公开）✅
            │    ├─ POST /login   → handlers.Auth.Login → authSvc.Login(bcrypt) → 写双 Cookie
            │    ├─ POST /refresh → handlers.Auth.Refresh → authSvc.RefreshToken → 写新 AT Cookie
            │    ├─ POST /logout  → handlers.Auth.Logout → 清除双 Cookie
            │    └─ GET  /me      → JWTAuth → handlers.Auth.Me → authSvc → repo.FindByID
            ├─ Admin 路由组（JWTAuth 中间件保护）✅
            │    ├─ handlers.Vehicles ✅ List/Create/Update/Delete/Publish/Unpublish/Duplicate/Batch
            │    │    ├─ vehicleSvc → VehicleRepo → *gorm.DB
            │    │    ├─ categorySvc → CategoryRepo → *gorm.DB（校验分类存在）
            │    │    └─ mediaRepo → *gorm.DB（拼 coverImageUrl）
            │    ├─ handlers.Categories ✅ List/Create/Update/Delete
            │    │    └─ categorySvc → CategoryRepo → *gorm.DB
            │    ├─ handlers.ParamTemplates ✅ Create/Update/Detail/AllDetail
            │    │    └─ paramTemplateSvc → ParamTemplateRepo → *gorm.DB（含事务 Update）
            │    ├─ handlers.System ✅ GET/POST/PUT /admin/system-settings
            │    │    └─ sysSvc → SysRepo → *gorm.DB
            │    └─ handlers.Upload ✅ POST /admin/upload/images
            │         ├─ oss.MinioClient PutObject
            │         └─ mediaRepo.Create → *gorm.DB
            └─ Public 路由组（无认证）✅ GET /vehicles → handlers.Vehicles.List
```

### 待办 / 断裂点

| # | 问题 | 文件 | 状态 |
|---|------|------|------|
| 1 | 车型列表/管理端 **分页** 仍由前端 slice（未走后端 page） | `handler/vehicles.go`、`admin` | ⏳ |
| 2 | `VehicleRepo` 无独立 **物理 Delete**（当前为 **逻辑删除** `status=deleted`） | 按产品决定是否补硬删 | ⏳ |
| 3 | 分类 `Update`/`Delete` | ✅ 已接通（见 [lesson-20260325](./lesson-20260325.md)） | ✅ |
| 4 | User 模块无 Service / Repository 层 | `handler/user.go`（仍为 stub） | ⏳ |
| 5 | 上传接口未挂载认证中间件 | `router/router.go` | ✅ 已由 admin 组 JWTAuth 中间件覆盖 |
| 6 | **media / OSS GC**（无引用对象清理） | 未实现 | ⏳ |
| 7 | 种子数据——初始管理员账号 | `cmd/seed/main.go` 或迁移 `000006` | ⏳ |
| 8 | `.env` 需补全 JWT 密钥 ≥ 32 字符 | `.env` / `.env.dev` | ⏳ |
| 9 | 公开端尚无系统设置/联系配置聚合接口 | `public/contact` 或 `public/home` 聚合 | ⏳ |
| 10 | 参数模板 **List / Delete** 路由未注册 | `handler/param_template.go`、`router.go` | ⏳ |
| 11 | `helper.RequiredField` 格式化 `%s` 缺参数 | `helper/utils.go` | ⏳ |

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
├── deploy/docker/      # 本地 PG/Redis/MinIO Compose ✅
├── migrations/         # SQL 迁移 ✅（000001～000005：init / categories / media_assets / system_settings / param_templates）
├── internal/bootstrap/ # 依赖装配 ✅（组合根：buildHandlers 集中 repo→service→handler）
├── internal/infrastructure/postgres/ # Postgres 连接 ✅
├── internal/infrastructure/oss/     # MinIO/S3 客户端 ✅
├── internal/domain/    # 领域语义 ✅（enum / model / rule）
├── internal/service/vehicle/ # 业务服务 ✅（GetById、写接口、Publish、Unpublish、Duplicate、Batch）
├── internal/service/system_setting/ # 系统设置业务 ✅（Detail、Create、Update）
├── internal/service/param_template/ # 参数模板业务 ✅（Create、Update、Detail、AllDetail）
├── internal/repository/postgres/ # 数据仓储 ✅（VehicleRepo、MediaAssetRepo、CategoryRepo、SystemRepo、ParamTemplateRepo）
├── internal/service/auth/    # 认证业务 ✅（Login、RefreshToken、GetCurrentUser）
├── internal/transport/http/
│   ├── router/         # 路由 ✅（收 Handlers 结构体，不碰 DB；auth 公开组 + admin JWTAuth 保护组 + public）
│   ├── handler/        # 处理器 ✅（只收 service/repo，不碰 *gorm.DB；车型、上传、auth、system-settings、param-templates）
│   ├── helper/         # 工具函数 ✅（分页解析、排序校验、RequiredField 泛型校验）
│   └── middleware/     # 中间件 ✅（JWTAuth、RequireRole、ValidateParams）
├── pkg/
│   ├── jwt/            # JWT 工具 ✅（Claims、GenerateToken、ParseToken）
│   ├── logger/         # 日志 ✅
│   └── response/       # 统一响应 ✅
└── ...
```

---

## 五、下一步建议（对齐 [循序渐进总说明](../循序渐进总说明.md) 第 6 步及后续）

1. **参数模板补全**：补 `GET /admin/param-templates` 列表分页 + `DELETE /admin/param-templates/:id` 路由注册。
2. **公开端系统设置**：补 `contact`/`home` 聚合读取，让小程序真正消费 `system_settings`（电话、微信、免责声明、默认分享图）。
3. **`.env` 补全 JWT 密钥**：`VEHIVLE_JWT_SECRET` 和 `VEHIVLE_JWT_REFRESH_SECRET` ≥ 32 字符，否则启动报错（Validate 已校验）。
4. **后端分页**：管理端车型列表改为 SQL `LIMIT/OFFSET` 或游标，与 admin 筛选对齐。
5. **公开 vs 管理端**：`List` 已用路径区分 `onlyPublished`；后续可改为显式参数或两套 handler，避免隐式依赖 URL。
6. **domain / rule**：按产品收紧 `Publish` 的详情图/参数校验（当前 MVP 可放宽）。
7. **种子数据**：创建初始管理员账号（`cmd/seed/main.go` 或迁移 `000006_seed_admin_user.up.sql`），需 bcrypt 哈希密码。
8. **（可选）** 优雅关闭时 `postgres.Close`；GORM SQL 日志接入统一 logger。
9. **（可选）** 媒体 GC：无引用 `media_assets` + 对应 OSS 对象清理任务。
10. **（V1.5）** Redis Token 黑名单（登出/踢人即时生效）。

---

## 六、每日进度

> 按日期倒序，最新在前。建议每天结束前：自测通过、写 5 行复盘、记录明天第一件事。

### 2026-04-13

- **完成**：**参数模板全链路落地**——从迁移到路由一招贯穿，新模块接入组合根只加 3 行
- **完成（迁移）**：新增 **`000005_create_param_templates_tables`**——`param_templates`（一级分类唯一绑定）+ `param_template_items`（字段定义：text/number/single_select）；联合唯一索引 `(template_id, field_key)`
- **完成（领域）**：`model.ParamTemplate`（含 `Items []ParamTemplateItem` 一对多关联）；`model.ParamBody`（更新 DTO，可选指针）；`model.ParamTemplateItem`（`*int8` 的 required/display 避免 bool 零值陷阱）
- **完成（枚举）**：`enum.ParamTemplateStatus`（int8 0/1） + `driver.Valuer` + `sql.Scanner`
- **完成（仓储）**：`ParamTemplateRepo`——Create / **Update（事务三步：更主表→upsert 子项→删已移除项）** / Delete / Detail / AllDetail(`Preload`) / Count
- **完成（服务）**：`ParamTemplateService`——Create / Update（先查存在性） / Detail / AllDetail
- **完成（接口）**：`POST/PUT/GET /admin/param-templates`（Create / Update / Detail / AllDetail）
- **完成（DI 接入）**：`buildHandlers()` 新增 `plateTemRepo → plateTemSvc → handler.NewParamTemplates(svc)`；`Handlers` 结构体新增 `ParamTemplates`
- **完成（helper）**：新增 `RequiredField[T ~string]` 泛型必填校验函数
- **完成（路由优化）**：404/405 错误信息中文化（含路径/方法提示）；注释清理
- **完成（管理端）**：参数模板页分类下拉仅显示一级分类（`level: 1`），label 改为"所属一级分类"
- **学习**：GORM 一对多 `Preload("Items")` 字段名 ≠ 列名；`Save()` 的 upsert 行为（有 PK 更新，无 PK 插入）
- **学习**：Go 泛型 `[T ~string]`——`~` 表示底层类型约束
- **学习**：事务内「更新 → upsert 子集 → 清理」模式适用于所有一对多编辑场景
- **文档**：新增 [lesson-20260413.md](./lesson-20260413.md)；更新 [README.md](./README.md)、[progress.md](./progress.md)
- **明日第一件事**：跑 `000005` 迁移 + 全流程联调参数模板 CRUD（含子项增删改）

### 2026-04-12

- **完成**：**依赖注入重构（组合根模式）**——把 repo→service→handler 的组装从 router/handler 搬到 `bootstrap.buildHandlers()`
- **完成（bootstrap）**：新增 `buildHandlers()` 方法，集中创建所有 repo、service、handler，返回 `router.Handlers` 结构体
- **完成（router）**：新增 `Handlers` 结构体聚合全部 handler；`Router` 去掉 `db *gorm.DB` 字段；构造函数 `New()` 不再需要 `db`；路由注册改为 `r.handlers.Xxx.Method`
- **完成（handler/auth）**：`NewAuth(db, jwtCfg)` → `NewAuth(authSvc, jwtCfg)`；去掉 `import postgres`、`import gorm`
- **完成（handler/user）**：`User { DB *gorm.DB }` → `User {}`；`NewUser(db)` → `NewUser()`（仍为 stub）
- **完成（handler/categories）**：`Categories { DB, CategoryService }` → `Categories { CategoryService }`；`NewCategories(db)` → `NewCategories(catSvc)`
- **完成（handler/vehicles）**：`NewVehicles(db, oss)` → `NewVehicles(vehSvc, catSvc, mediaRepo, oss)`
- **完成（handler/system）**：`NewSysSettings(db, oss)` → `NewSysSettings(sysSvc, oss)`；去掉 `DB` 字段
- **完成（handler/upload）**：`Upload { OSS, DB }` → `Upload { OSS, mediaRepo }`；`u.DB.WithContext(ctx).Create(row)` → `u.mediaRepo.Create(ctx, row)`
- **学习**：组合根（Composition Root）——应用中「唯一知道所有具体实现」的地方；Go 的依赖注入靠构造函数传参，简单、显式、零魔法
- **学习**：`*gorm.DB` 只应出现在 bootstrap（创建连接）和 repository（执行查询）两层，中间层不碰
- **学习**：Handler 结构体字段应「诚实」——没用到的依赖不挂（如 `Categories.DB` 从未被方法调用）
- **文档**：新增 [lesson-20260412.md](./lesson-20260412.md)（含完整调用链路图、前后端 DI 对照表）；更新 [README.md](./README.md)、[progress.md](./progress.md)
- **明日第一件事**：启动服务跑一遍 CRUD 接口确认行为不变（功能回归）

### 2026-04-11

- **完成**：**系统设置单行表落地**——`system_settings` 从旧 key-value/JSONB 结构重构为**单行配置表**
- **完成（迁移）**：新增 **`000004_alter_system_settings`**，字段固定为 `company_name`、客服电话/微信、`default_price_mode`、免责声明、默认分享标题/图；`id = 1` + `CHECK` 保证全局仅一行
- **完成（领域）**：新增 `model.SystemSetting`，必填字段用值类型、可空字段用 `*string`
- **完成（仓储）**：新增 `SystemRepo`：`Detail` / `Create` / `Update` / `Exists`，写操作固定作用于 `id = 1`
- **完成（服务）**：新增 `system_setting.SysService`，收口“首次创建、后续更新”的业务规则：已存在不可重复创建，不存在不可更新
- **完成（接口）**：新增 **`GET/POST/PUT /api/v1/admin/system-settings`**；详情无数据时返回 `data: null`
- **完成（响应适配）**：系统设置响应补充 `defaultShareImageUrl`，由 `storage_key` 通过 OSS 客户端拼接预览 URL
- **完成（管理端）**：新增系统设置页；支持公司名称、客服电话、客服微信、默认价格模式、免责声明、默认分享标题/图编辑
- **完成（映射）**：管理端 `defaultPriceMode` 做前后端枚举转换：`msrp ↔ show_price`、`negotiable ↔ phone_inquiry`
- **学习**：单行表比 key-value 更适合“全局唯一配置对象”；`NULL` 与空字符串不是同一语义；`storage_key`（存储层）与 `public URL`（展示层）应分离
- **文档**：新增 [lesson-20260411.md](./lesson-20260411.md)；更新 [README.md](./README.md)、[progress.md](./progress.md) 索引与全链路
- **明日第一件事**：补公开端 `system_settings` / `contact` 聚合读取接口，打通小程序联系方式、免责声明、默认分享图兜底

### 2026-04-10

- **完成**：**JWT 认证闭环落地**——对照 [jwt-implementation-guide.md](../jwt-implementation-guide.md) 全链路审计，发现并修复 **9 个偏离/断裂点**
- **修复（安全）**：`AdminUser.PasswordHash` 的 `json:"passwordHash"` → `json:"-"`，阻止密码哈希序列化到响应
- **修复（安全）**：Login 错误信息统一为 `ErrInvalidCredentials`（"用户名或密码错误"），不泄露具体原因
- **修复（安全）**：Cookie `SameSite` 从 `NoneMode` 改为 `LaxMode`，防 CSRF
- **修复（安全）**：`configs.validateJWTRequired()`——密钥必填、≥ 32 字符、生产强制 CookieSecure
- **修复（功能）**：实现 **Refresh 端点**（`Service.RefreshToken` + `Handler.Refresh` + 路由注册），从 RT Cookie 读取并签发新 AT
- **修复（功能）**：RT Cookie `MaxAge` 使用独立的 `RefreshExpiresIn`（之前误用 AT 有效期）
- **修复（功能）**：Login 响应体返回 `{ expiresIn }` 供前端做续签倒计时
- **修复（功能）**：Login 失败错误码从 `CodeBusinessError` 改为 `FailAuth`（`A00001`）
- **修复（功能）**：`FailAuth` 从 HTTP 200 改为 **HTTP 401**、`FailAuthDenied` 改为 **HTTP 403**——**打通前端无感刷新链路**（Axios error 拦截器仅在非 2xx 时触发）
- **新增**：`RequireRole` 角色鉴权中间件，`map[string]bool` O(1) 查找
- **新增**：`JWTConfig` 扩展 `CookieDomain` + `CookieSecure` 字段
- **清理**：删除 `bootstrap.jwtConnPool()`（冗余，已由 `Validate` 覆盖）
- **学习**：`json:"-"` 序列化跳过、统一错误信息防信息泄露、`SameSite=Lax` 与 CSRF、RT Cookie Path 限制、Refresh Token 续签流程、`RequireRole` 中间件工厂函数闭包、HTTP 状态码与 Axios 拦截器的联动关系（200 不触发 error 回调）
- **文档**：新增 [lesson-20260410.md](./lesson-20260410.md)；更新 [README.md](./README.md)、[progress.md](./progress.md) 全链路
- **明日第一件事**：种子数据（初始管理员账号）+ `.env` 补全 JWT 密钥 ≥ 32 字符 → 自测四接口

### 2026-04-09

- **完成**：迁移 **`000003`** —— 表 **`media_assets`**（`storage_key` UNIQUE、`asset_type`、`created_at`）
- **完成**：**上传后写库** —— `PutObject` 成功后 `INSERT media_assets`，响应 **`{ id, url, storageKey }`**；`NewUpload(oss, db)`
- **完成**：**领域** —— `model.MediaAsset`；`Vehicle.CoverImageURL`（`gorm:"-"`）；`Vehicle` 列 tag + `TableName()`
- **完成**：**MediaAssetRepo** —— `MapStorageKeysByIDs` 批量补全封面 **`storage_key`**
- **完成**：**OSS** —— `MinioClient.ObjectPublicURL` 与上传 URL 规则一致
- **完成**：**VehicleService** —— `GetById`、`VehicleUpdateInput`、`UpdateVehicle`、`SoftDelete`、`Unpublish`、`Duplicate`、`BatchSetStatus`
- **完成**：**Vehicles Handler** —— `attachCoverImageURLs`；`Update`/`Delete`/`Publish`/`Unpublish`/`Duplicate`/`BatchStatus`；`NewVehicles(db, oss)`
- **完成**：**Router** —— `POST batch-status`、`publish`、`unpublish`、`duplicate`；`public` 与 admin 共用 `NewVehicles(db, oss)`
- **完成**：**前端** —— `ImageUploadResult`；`coverMediaId` = 媒体 **UUID**；`previewFromServer` 回显
- **文档**：新增 [lesson-20260409.md](./lesson-20260409.md)；更新 [README.md](./README.md)、[progress.md](./progress.md) 全链路
- **明日第一件事**：上传鉴权 JWT；或后端车型分页

### 2026-04-08

- **完成**：**MinIO/S3 对象存储集成** —— `infrastructure/oss/client.go`（`MinioClient` 结构体，聚合 endpoint、bucket、`*minio.Client`）
- **完成**：**Bootstrap 生命周期扩展** —— 新增 `ossConnPool()`（MinIO 连接初始化 + Bucket 存在性检查 + 自动创建）；启动顺序：cfg → logger → DB → Ping → **OSS** → Router
- **完成**：**`normalizeOssEndpoint`** —— 将用户配置的 URL（`http://...` / `https://...` / 裸 `host:port`）统一归一化为 MinIO SDK 所需的 `host:port` + TLS 标志
- **完成**：**图片上传 Handler** —— `handler/upload.go`（MIME 白名单、10MB 大小限制、`images/<date>/<uuid>` 唯一对象键、`PutObject` 写入 MinIO、返回公开 URL）
- **完成**：**Router 扩展** —— `Router` 结构体新增 `oss` 字段，注册 `POST /admin/upload/images`；健康检查返回 `ossReady`
- **完成**：**配置层增强** —— `OssConfig` 新增 `Endpoint`/`UseSSL`；`validateOssRequired()` 启动前校验五项必填；`default.go` 新增 `DefaultOSSRegion`/`DefaultOSSBucket`
- **完成**：**Docker Compose 扩展 MinIO** —— 新增 `minio` 服务（S3 API:9000 + Console:9001）、`vehivle_minio_data` 持久卷、healthcheck
- **完成**：**统一响应 `RequestID` 补全** —— `Success`/`Fail`/`FailNotFound`/`FailMethodNotAllowed` 均填入 `getRequestID(c)`
- **完成**：**`injectRouter` 错误处理改进** —— 返回 `error` 替代 `os.Exit(1)`，遵循 Go 错误处理惯例
- **完成**：**前端适配** —— `mediaApi` 从两步上传（policy+complete）简化为单步直传 `uploadImage(file)`；`ImageUploader` 组件同步更新；`coverMediaId` 语义改为 OSS 对象键
- **完成**：**代码规范** —— 清理冗余注释；错误信息中文化；`CategoryRepo` 接口方法添加文档注释
- **学习**：`defer` 关键字（Go 的 `try...finally`）、`10 << 20` 位运算表达 10MB、`%w` 错误包装、`map[string]bool` 做白名单 O(1) 查找
- **学习**：MinIO SDK `minio.New` 的 endpoint 要求（host:port 无 scheme）；`BucketExists`/`MakeBucket` 幂等模式；`PutObject` 参数（`io.Reader`）
- **学习**：Go struct 零值陷阱——`string` 零值 `""` 不报错但可能静默出错（如 `RequestID` 漏赋值）
- **文档**：新增 [lesson-20260408.md](./lesson-20260408.md)，更新 [learn/README.md](./README.md) 索引；[progress.md](./progress.md) 全链路、目录结构同步
- **明日第一件事**：上传接口挂载认证中间件，或车型 `Update`/`Delete` 接 Service + Repo

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

*最后更新：2026-04-13（参数模板全链路——000005 迁移 + model + repo(事务) + service + handler + 组合根接入；lesson-20260413 / README / progress 同步）*
