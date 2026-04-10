# JWT 认证闭环实现指导文档

> **对齐**：[循序渐进总说明](./循序渐进总说明.md) **第 5 步——做认证闭环**。  
> **前置**：第 1～4.5 步已完成（工程壳、HTTP 基座、Domain、数据库迁移、对象存储）。  
> **数据库**：`admin_users` 表已在 `000001_init_schema.up.sql` 中建好（UUID 主键、`username`、`password_hash`、`role`）。  
> **配置**：`configs.JWTConfig` 已定义 `Secret` + `ExpireHours`，`.env.example` 预留 `VEHIVLE_JWT_SECRET`。  
> **响应码**：`A00001`（认证失败）、`A00004`（授权失败）及 `FailAuth`/`FailAuthDenied` 已就绪。

| 元数据 | 值 |
|--------|-----|
| **涉及目录** | `pkg/jwtutil`、`domain/model`、`repository/postgres`、`service/auth`、`transport/http/handler`、`transport/http/middleware`、`bootstrap`、`configs` |
| **新增依赖** | `github.com/golang-jwt/jwt/v5`、`golang.org/x/crypto/bcrypt`（显式引入） |
| **对应技术文档** | [doc/tech.md](../../doc/tech.md) §3.3 管理后台接口：login / refresh / me |

---

## 一、认证方案选型

### 1.1 为什么选 JWT 而不是 Session

| 对比项 | Session | JWT |
|--------|---------|-----|
| 状态存储 | 服务端 Redis/内存 | Token 自包含，服务端无状态 |
| 水平扩展 | 需共享 Session Store | 天然可扩展，任意节点验签即可 |
| 传输载体 | Cookie（Session ID） | Cookie 或 Authorization Header 均可 |
| 适用场景 | 传统 SSR | 前后端分离、SPA、微服务 |

本项目为 **React SPA + Go API 前后端分离架构**，JWT 是企业级标准选择。

### 1.3 Token 传输方式：httpOnly Cookie（而非 Authorization Header）

企业实践中 JWT 的传输方式有两种主流选择：

| 对比项 | Authorization Header（Bearer） | httpOnly Cookie |
|--------|-------------------------------|-----------------|
| XSS 防护 | JS 可从 localStorage 读取 Token，**有泄露风险** | httpOnly 标记后 JS 完全无法读取，**天然免疫 XSS** |
| CSRF 防护 | 天然免疫（不自动携带） | 浏览器自动携带 Cookie，需配合 `SameSite` 防护 |
| 前端复杂度 | 需手动管理 Token 存储、Axios 拦截器注入 Header | 浏览器自动携带，**前端零管理** |
| 登出 | 前端清除 localStorage | 服务端 `Set-Cookie: MaxAge=0` 清除，**更可靠** |
| 移动端兼容 | 原生 App 友好 | WebView 可用，原生 App 需额外处理 |

**本项目选择 httpOnly Cookie**，理由：
1. 管理后台是纯 Web 应用，不涉及原生 App 调用
2. httpOnly Cookie 对 XSS 的防护是架构级的，不依赖前端代码质量
3. 前端无需手动管理 Token 生命周期，降低对接复杂度
4. 配合 `SameSite=Lax` 可在 V1 阶段有效防御 CSRF，无需额外 Token

### 1.2 单 Token vs 双 Token

技术文档写的是「JWT 单 Token」，但企业实践中推荐 **Access Token + Refresh Token 双 Token 模式**：

| 方案 | Access Token 有效期 | Refresh Token 有效期 | 优势 | 劣势 |
|------|---------------------|----------------------|------|------|
| 单 Token | 较长（24h） | 无 | 简单 | 过期即重新登录；Token 泄露风险窗口大 |
| 双 Token | 较短（2h） | 较长（7d） | AT 泄露窗口小；用户无感续签 | 多一次 refresh 请求 |

**建议 V1 MVP 采用双 Token**。即使 V1 先不接 Redis 黑名单，双 Token 的短有效期本身就是一层安全兜底。后续 V1.5 接 Redis 后可叠加「登出吊销」能力。

---

## 二、整体流程图

```
┌─────────────┐                           ┌──────────────────────────┐
│  React 管理后台 │                           │      Go Server           │
│              │                           │                          │
│  1. 登录表单  │── POST /auth/login ──────▶│  验证 username+password  │
│              │◀─ Set-Cookie:             │  bcrypt 比对              │
│              │    access_token=xxx;       │  签发 AT + RT             │
│              │    HttpOnly; SameSite=Lax  │  写入 httpOnly Cookie    │
│              │   Set-Cookie:             │                          │
│              │    refresh_token=xxx;      │                          │
│              │    HttpOnly; SameSite=Lax  │                          │
│              │                           │                          │
│  2. 请求 API  │── Cookie:                 │  JWT 中间件               │
│  （浏览器自动  │   access_token=xxx ──────▶│  从 Cookie 读取 AT       │
│    携带 AT）  │◀─ 正常业务响应 ────────────│  解析 → 验签 → 注入 Claims │
│              │                           │                          │
│  3. AT 过期   │── POST /auth/refresh ────▶│  从 Cookie 读取 RT       │
│  （浏览器自动  │◀─ Set-Cookie:             │  验证 RT 有效性           │
│    携带 RT）  │    access_token=新xxx;     │  签发新 AT → 写入 Cookie │
│              │    HttpOnly; SameSite=Lax  │                          │
│              │                           │                          │
│  4. 获取用户  │── Cookie: access_token ──▶│  从 Cookie 读取 AT       │
│  （需 AT）    │◀─ { id, username, role } ─│  解析 Claims → 查库      │
│              │                           │                          │
│  5. 登出     │── POST /auth/logout ─────▶│  Set-Cookie: MaxAge=0    │
│              │◀─ 两个 Cookie 被清除 ──────│  AT + RT Cookie 均过期   │
└─────────────┘                           └──────────────────────────┘
```

**关键区别**：Token 不出现在 JSON 响应体中，全程通过 `Set-Cookie` / `Cookie` 头传输，前端 JS 无法读取（httpOnly），浏览器自动管理。

---

## 三、新增依赖

```bash
cd server

# JWT 签发与解析
go get github.com/golang-jwt/jwt/v5

# 密码哈希（已是间接依赖，显式引入以在业务中使用）
go get golang.org/x/crypto
```

> **为什么选 `golang-jwt/jwt/v5`**：这是 `dgrijalva/jwt-go` 的官方继任维护库，Go 社区事实标准，GitHub 14k+ star，支持泛型 Claims、严格算法校验。

---

## 四、新增文件清单与职责

按项目已有分层约定，需新增以下文件：

```
server/
├── pkg/
│   └── jwtutil/
│       └── jwtutil.go              ← 纯函数：Token 签发、解析、Claims 定义
├── internal/
│   ├── domain/model/
│   │   └── admin_user.go           ← AdminUser 领域模型（对应 admin_users 表）
│   ├── repository/postgres/
│   │   └── admin_user_repo.go      ← 用户数据访问（FindByUsername、FindByID）
│   ├── service/auth/
│   │   └── service.go              ← 认证业务：Login、Refresh、GetCurrentUser
│   └── transport/http/
│       ├── handler/
│       │   └── auth.go             ← HTTP 入口：Login、Refresh、Me、Logout（写/清 Cookie）
│       └── middleware/
│           └── jwt_auth.go         ← JWT 认证中间件 + 角色鉴权中间件
├── configs/
│   └── conf.go                     ← 修改：JWTConfig 扩展、Validate 增加 JWT 校验
└── internal/bootstrap/
    └── bootstrap.go                ← 修改：传递 cfg 到 Router
```

### 改动文件清单

| 文件 | 改动类型 | 说明 |
|------|---------|------|
| `configs/conf.go` | 修改 | `JWTConfig` 增加 `RefreshExpireHours`；`Validate()` 增加密钥校验 |
| `configs/default.go` | 修改 | 增加 `DefaultJWTRefreshExpireHours` 常量 |
| `internal/bootstrap/bootstrap.go` | 修改 | `injectRouter` 传入 `cfg` |
| `internal/transport/http/router/router.go` | 修改 | Router 持有 `cfg`；注册 auth 路由组；admin 组挂载 JWT 中间件 |
| `handler/user.go` | 删除/重构 | 当前为 stub，登录逻辑移入 `handler/auth.go` |

---

## 五、各层实现指导

### 5.1 `pkg/jwtutil/jwtutil.go` — JWT 工具包

**定位**：纯函数层，不依赖任何业务包。放在 `pkg/` 是因为它与前端工具函数类似——可被任何内部包复用。

**核心职责**：
1. 定义 `Claims` 结构体（嵌入 `jwt.RegisteredClaims`）
2. `GenerateToken(secret, duration, userID, username, role) → (tokenString, error)` 签发
3. `ParseToken(secret, tokenString) → (*Claims, error)` 解析验签

**企业规范要点**：

| 编号 | 规范 | 原因 |
|------|------|------|
| 1 | 签名算法固定 `HS256`，解析时严格校验 `SigningMethod` | 防止 `alg: none` 攻击（CVE-2015-9235） |
| 2 | `Issuer` 写死 `"vehivle-admin"` | 多服务场景可区分 Token 来源 |
| 3 | 错误区分「过期」和「无效」 | 前端据此走不同逻辑（过期 → refresh；无效 → 重新登录） |
| 4 | 不在 Claims 里放敏感信息 | JWT payload 是 Base64 编码，**不是加密**，任何人可解码查看 |

**Claims 应包含的字段**：

```
Claims {
    UserID   string   // 用户 UUID，用于后续查库
    Username string   // 用户名，日志/审计用
    Role     string   // 角色标识，中间件鉴权用
    jwt.RegisteredClaims {
        ExpiresAt  // 过期时间
        IssuedAt   // 签发时间
        NotBefore  // 生效时间
        Issuer     // 签发方
    }
}
```

**不应放入 Claims 的字段**：密码哈希、手机号、邮箱等 PII（个人可识别信息）。

---

### 5.2 `internal/domain/model/admin_user.go` — 领域模型

**对齐数据库**：与 `migrations/000001_init_schema.up.sql` 中的 `admin_users` 表一一对应。

**关键约定**：

| 字段 | GORM Tag | JSON Tag | 说明 |
|------|----------|----------|------|
| `ID` | `column:id;type:uuid;primaryKey` | `json:"id"` | UUID 主键 |
| `Username` | `column:username;uniqueIndex;size:64` | `json:"username"` | 登录名 |
| `PasswordHash` | `column:password_hash` | `json:"-"` | **必须 `"-"`，绝不序列化到响应** |
| `Role` | `column:role;size:32;default:editor` | `json:"role"` | `super_admin` / `editor` |
| `CreatedAt` | `autoCreateTime` | `json:"createdAt"` | 创建时间 |
| `UpdatedAt` | `autoUpdateTime` | `json:"updatedAt"` | 更新时间 |

**必须实现 `TableName() string`**：返回 `"admin_users"`，与迁移表名对齐，避免 GORM 默认复数推导不一致。

> 类比前端：`json:"-"` 相当于 TypeScript 接口中不导出某个字段，确保 `JSON.stringify` 时不会泄露密码哈希。

---

### 5.3 `internal/repository/postgres/admin_user_repo.go` — 数据访问

**需要的方法**（参考已有的 `VehicleRepo` 和 `CategoryRepo` 模式）：

| 方法 | 用途 | 调用场景 |
|------|------|---------|
| `FindByUsername(ctx, username) → (*AdminUser, error)` | 按用户名精确查询 | 登录 |
| `FindByID(ctx, id) → (*AdminUser, error)` | 按 UUID 查询 | Token 解析后获取用户详情（`/auth/me`） |

**注意**：
- 方法参数带 `context.Context`，与已有 Repo 风格保持一致
- 使用 `db.WithContext(ctx)` 传递链路上下文
- 登录场景不需要返回列表，只查单条

---

### 5.4 `internal/service/auth/service.go` — 认证业务服务

这是认证的**核心编排层**，职责：

| 方法 | 职责 | 安全规范 |
|------|------|---------|
| `Login(ctx, username, password)` | 查用户 → bcrypt 比对 → 签发双 Token | 失败统一返回「用户名或密码错误」，不泄露具体原因 |
| `RefreshToken(ctx, refreshTokenString)` | 解析 RT → 验证有效性 → 签发新 AT | RT 过期则要求重新登录 |
| `GetCurrentUser(ctx, userID)` | 按 ID 查用户信息 | Token 已验证，直接查库 |

**Login 流程详解**：

```
Login(username, password)
  │
  ├─ 1. repo.FindByUsername(username)
  │     └─ 未找到 → return ErrInvalidCredentials（不说「用户不存在」）
  │
  ├─ 2. bcrypt.CompareHashAndPassword(user.PasswordHash, password)
  │     └─ 不匹配 → return ErrInvalidCredentials（不说「密码错误」）
  │
  ├─ 3. jwtutil.GenerateToken(secret, atDuration, user.ID, user.Username, user.Role)
  │     └─ 签发 Access Token（短有效期，如 2 小时）
  │
  ├─ 4. jwtutil.GenerateToken(secret, rtDuration, user.ID, user.Username, user.Role)
  │     └─ 签发 Refresh Token（长有效期，如 7 天）
  │
  └─ 5. return { accessToken, refreshToken, expiresIn }
        （Token 字符串返回给 Handler，由 Handler 写入 Cookie；Service 不感知传输方式）
```

**密码哈希说明**：

- 使用 `golang.org/x/crypto/bcrypt`，**不要用** MD5/SHA256
- bcrypt 自带盐值，`cost=10`（默认值）在当前硬件下每次哈希约 100ms，足以防暴力破解
- `CompareHashAndPassword` 是**恒定时间比较**，防时序攻击

> 类比前端：bcrypt 类似于前端「不要用 `==` 比较密码」的理念升级版——它不仅比较安全，还自带加盐和慢哈希抗暴力破解。

---

### 5.5 `internal/transport/http/middleware/jwt_auth.go` — 认证中间件

**两个中间件**：

#### （1）`JWTAuth(secret string) gin.HandlerFunc` — JWT 认证

Token 从 **Cookie** 中读取（而非 Authorization Header），浏览器自动携带，前端零管理。

执行流程：

```
请求进入
  │
  ├─ 1. 从 Cookie 读取 access_token
  │     c.Cookie("access_token")
  │     └─ Cookie 不存在或为空 → FailAuth("缺少有效的认证令牌") + Abort
  │
  ├─ 2. jwtutil.ParseToken(secret, tokenString)
  │     ├─ ErrTokenExpired → FailAuth("认证令牌已过期") + Abort
  │     └─ ErrTokenInvalid → FailAuth("认证令牌无效") + Abort
  │
  ├─ 3. 将 Claims 注入 gin.Context（供后续 Handler 读取）
  │     ├─ c.Set("current_user_id", claims.UserID)
  │     ├─ c.Set("current_username", claims.Username)
  │     └─ c.Set("current_role", claims.Role)
  │
  └─ 4. c.Next()
```

**与 Bearer Token 方案的差异**：不需要解析 `Authorization` Header、不需要去掉 `Bearer ` 前缀——直接 `c.Cookie("access_token")` 一步到位。

**Cookie Name 建议定义为常量**：

```
CookieAccessToken  = "access_token"
CookieRefreshToken = "refresh_token"
CtxKeyUserID       = "current_user_id"
CtxKeyUsername     = "current_username"
CtxKeyRole         = "current_role"
```

#### （2）`RequireRole(allowedRoles ...string) gin.HandlerFunc` — 角色鉴权

- 从 `c.Get(CtxKeyRole)` 读取当前角色
- 不在允许列表中 → `FailAuthDenied("无权访问该接口")` + `Abort`
- V1 角色仅 `super_admin` 和 `editor`，大多数接口两者都可用
- 仅特定危险操作（如删除用户、修改系统设置）限制为 `super_admin`

---

### 5.6 `internal/transport/http/handler/auth.go` — Auth Handler

**四个接口**（对齐 tech.md §3.3，新增 logout）：

| 接口 | 方法 | 路径 | 是否需要 JWT | Cookie 行为 |
|------|------|------|-------------|------------|
| 登录 | POST | `/api/v1/admin/auth/login` | 否（公开） | 写入 AT + RT Cookie |
| 刷新 | POST | `/api/v1/admin/auth/refresh` | 否（浏览器自动携带 RT Cookie） | 写入新 AT Cookie |
| 当前用户 | GET | `/api/v1/admin/auth/me` | 是（浏览器自动携带 AT Cookie） | 无 |
| 登出 | POST | `/api/v1/admin/auth/logout` | 否 | 清除 AT + RT Cookie |

**请求/响应契约**：

```
# POST /api/v1/admin/auth/login
Request Body:
  { "username": "admin", "password": "123456" }

Response Headers (成功):
  Set-Cookie: access_token=eyJhbGci...; Path=/api; HttpOnly; SameSite=Lax; Max-Age=7200
  Set-Cookie: refresh_token=eyJhbGci...; Path=/api/v1/admin/auth; HttpOnly; SameSite=Lax; Max-Age=604800

Response Body (成功):
  {
    "code": "000000",
    "message": "success",
    "data": {
      "expiresIn": 7200
    }
  }

Response Body (失败):
  {
    "code": "A00001",
    "message": "用户名或密码错误"
  }
```

> **注意**：Token 不在 JSON body 中返回，仅通过 `Set-Cookie` 头写入。`data` 中只返回 `expiresIn` 供前端做倒计时续签。

```
# POST /api/v1/admin/auth/refresh
Request: 无 body（浏览器自动携带 refresh_token Cookie）

Response Headers (成功):
  Set-Cookie: access_token=eyJhbGci...; Path=/api; HttpOnly; SameSite=Lax; Max-Age=7200

Response Body (成功):
  {
    "code": "000000",
    "data": {
      "expiresIn": 7200
    }
  }
```

> **Refresh Token 的 Cookie Path 限制**：RT Cookie 的 `Path` 设为 `/api/v1/admin/auth`，使其**仅在 auth 路径下被浏览器携带**，不会随普通业务请求发送，缩小泄露面。

```
# GET /api/v1/admin/auth/me
Cookie: access_token=eyJhbGci...（浏览器自动携带）

Response Body:
  {
    "code": "000000",
    "data": {
      "id": "uuid-xxx",
      "username": "admin",
      "role": "super_admin"
    }
  }
```

```
# POST /api/v1/admin/auth/logout
Response Headers:
  Set-Cookie: access_token=; Path=/api; HttpOnly; SameSite=Lax; Max-Age=0
  Set-Cookie: refresh_token=; Path=/api/v1/admin/auth; HttpOnly; SameSite=Lax; Max-Age=0

Response Body:
  {
    "code": "000000",
    "message": "success"
  }
```

**Cookie 写入工具函数建议**：

Handler 中频繁调用 `c.SetCookie()`，建议抽取为辅助函数统一管理 Cookie 属性：

```
setAccessTokenCookie(c, token, maxAge)   → c.SetCookie("access_token", token, maxAge, "/api", "", secure, true)
setRefreshTokenCookie(c, token, maxAge)  → c.SetCookie("refresh_token", token, maxAge, "/api/v1/admin/auth", "", secure, true)
clearAuthCookies(c)                      → 两个 Cookie 均设 MaxAge=0
```

**`c.SetCookie` 各参数含义**：

| 参数 | 值 | 说明 |
|------|-----|------|
| name | `"access_token"` / `"refresh_token"` | Cookie 名称 |
| value | Token 字符串 / `""` | Cookie 值（清除时为空串） |
| maxAge | 秒数 / `0` | 有效期（0 = 会话级，-1 或 0 用于删除） |
| path | `"/api"` / `"/api/v1/admin/auth"` | Cookie 生效路径范围 |
| domain | `""` | 空串表示当前域（本地开发 `localhost`） |
| secure | `cfg.App.Env == "prod"` | 生产 true（仅 HTTPS），开发 false |
| httpOnly | `true` | **必须 true**，JS 不可读 |

---

### 5.7 路由改造 — `router/router.go`

核心变更：auth 路由独立为**无需认证的公开组**，其余 admin 路由挂载 JWT 中间件。

**改造前**（当前结构）：

```
/api/v1
  └── /admin          ← 无任何认证
        ├── /user
        ├── /vehicles
        ├── /categories
        └── /upload/images
```

**改造后**（目标结构）：

```
/api/v1
  ├── /admin/auth               ← 公开（无 JWT 中间件）
  │     ├── POST /login           写入 AT + RT Cookie
  │     ├── POST /refresh         从 RT Cookie 读取，写入新 AT Cookie
  │     └── POST /logout          清除 AT + RT Cookie
  │
  ├── /admin                    ← 受保护（挂载 JWTAuth 中间件，从 AT Cookie 读取）
  │     ├── GET  /auth/me
  │     ├── /vehicles/...
  │     ├── /categories/...
  │     └── POST /upload/images
  │
  └── /public                   ← 公开（保持不变）
        └── GET /vehicles
```

**Router 结构体需要扩展**：新增 `cfg *configs.Conf` 字段，以便中间件获取 `JWT.Secret`，Handler 获取 Cookie 配置。

**新增 logout 路由**：`POST /admin/auth/logout` 放在 auth 公开组（不依赖 JWT 中间件，服务端只需清除 Cookie）。

**依赖注入链路变更**：

```
Bootstrap
  └── injectRouter(gin, logger, db, oss, cfg)   ← 新增 cfg
        └── Router.Register()
              ├── NewAdminUserRepo(db)
              ├── auth.NewService(repo, cfg.JWT)
              ├── handler.NewAuth(authService, cfg)  ← Handler 需要 cfg 以写 Cookie
              └── admin.Use(middleware.JWTAuth(cfg.JWT.Secret))
```

---

### 5.8 配置层变更 — `configs/`

#### （1）`JWTConfig` 扩展

当前：

```go
type JWTConfig struct {
    Secret      string `mapstructure:"secret"`
    ExpireHours int    `mapstructure:"expire_hours"`
}
```

建议扩展为：

```go
type JWTConfig struct {
    Secret             string `mapstructure:"secret"`
    ExpireHours        int    `mapstructure:"expire_hours"`          // AT 有效期（小时）
    RefreshExpireHours int    `mapstructure:"refresh_expire_hours"`  // RT 有效期（小时）
    CookieDomain       string `mapstructure:"cookie_domain"`        // Cookie 域名（本地留空，生产填域名）
    CookieSecure       bool   `mapstructure:"cookie_secure"`        // 仅 HTTPS 传输（生产 true）
}
```

#### （2）`default.go` 新增默认值

```
DefaultJWTExpireHours        = 2       // AT 2 小时
DefaultJWTRefreshExpireHours = 168     // RT 7 天（7 × 24）
DefaultJWTCookieDomain       = ""      // 空 = 当前域（localhost 开发适用）
DefaultJWTCookieSecure       = false   // 开发环境 HTTP，生产走 HTTPS 时设 true
```

#### （3）`Validate()` 增加 JWT 校验

```go
// 启动时必须配置密钥
if strings.TrimSpace(c.JWT.Secret) == "" {
    return fmt.Errorf("jwt.secret is required (set VEHIVLE_JWT_SECRET)")
}
// HS256 密钥长度建议 ≥ 32 字符
if len(c.JWT.Secret) < 32 {
    return fmt.Errorf("jwt.secret must be at least 32 characters for HS256 security")
}
// 生产环境强制 Cookie Secure
if c.App.Env == "prod" && !c.JWT.CookieSecure {
    return fmt.Errorf("jwt.cookie_secure must be true in production (HTTPS only)")
}
```

#### （4）环境变量配置

```env
# .env（本地开发）
VEHIVLE_JWT_SECRET=your-development-secret-at-least-32-chars
VEHIVLE_JWT_EXPIRE_HOURS=2
VEHIVLE_JWT_REFRESH_EXPIRE_HOURS=168
VEHIVLE_JWT_COOKIE_DOMAIN=
VEHIVLE_JWT_COOKIE_SECURE=false
```

```env
# .env.prod（生产环境）
VEHIVLE_JWT_SECRET=<运维生成的随机密钥>
VEHIVLE_JWT_EXPIRE_HOURS=2
VEHIVLE_JWT_REFRESH_EXPIRE_HOURS=168
VEHIVLE_JWT_COOKIE_DOMAIN=yourdomain.com
VEHIVLE_JWT_COOKIE_SECURE=true
```

> **重要**：`jwt.secret` 只走环境变量，**永远不提交到 YAML 文件或代码仓库**。

#### （5）Cookie 属性速查表

| 属性 | AT Cookie | RT Cookie | 说明 |
|------|-----------|-----------|------|
| Name | `access_token` | `refresh_token` | 固定常量 |
| Path | `/api` | `/api/v1/admin/auth` | RT 限制在 auth 路径，缩小携带范围 |
| Domain | `cfg.JWT.CookieDomain` | 同左 | 本地留空，生产填顶级域名 |
| HttpOnly | `true` | `true` | **必须**，JS 不可读 |
| Secure | `cfg.JWT.CookieSecure` | 同左 | 生产 true（仅 HTTPS） |
| SameSite | `Lax` | `Lax` | 防 CSRF，允许顶级导航携带 |
| Max-Age | `ExpireHours × 3600` | `RefreshExpireHours × 3600` | 秒数 |

---

## 六、种子数据——初始管理员账号

数据库中需要一条初始管理员记录才能登录。有两种方式：

### 方案 A：SQL 迁移种子（推荐 V1）

新建迁移文件 `000004_seed_admin_user.up.sql`：

```sql
-- 种子数据：初始超级管理员账号
-- 密码为 bcrypt 哈希，明文不落库
-- 默认密码: Admin@2026（上线后必须修改）
INSERT INTO admin_users (username, password_hash, role)
VALUES (
    'admin',
    '$2a$10$xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx',
    'super_admin'
) ON CONFLICT (username) DO NOTHING;
```

**生成 bcrypt 哈希的方法**（在项目内创建一个临时脚本或用命令行）：

```bash
# 方法 1：Go 一行脚本
go run -e 'package main; import ("fmt"; "golang.org/x/crypto/bcrypt"); func main() { h, _ := bcrypt.GenerateFromPassword([]byte("Admin@2026"), 10); fmt.Println(string(h)) }'

# 方法 2：写一个 cmd/seed/main.go 命令行工具（推荐，可复用）
```

### 方案 B：Seed 命令行工具（更规范）

在 `cmd/seed/main.go` 中编写种子数据初始化逻辑，支持幂等执行。这与 `cmd/migrate/main.go` 同级，符合项目已有的 CLI 工具约定。

---

## 七、安全规范清单

| 编号 | 规范 | 理由 | 检查方式 |
|------|------|------|---------|
| S01 | JWT Secret ≥ 32 字节 | HS256 需足够熵值 | `Validate()` 启动时校验 |
| S02 | Secret 只走环境变量 | 防止泄露到代码仓库 | `.gitignore` 含 `.env`；YAML 中 `jwt.secret` 留空 |
| S03 | bcrypt 哈希密码（cost ≥ 10） | 慢哈希 + 自带盐，防暴力破解 | 代码 review |
| S04 | 登录失败不泄露原因 | 统一「用户名或密码错误」 | 接口测试 |
| S05 | Token 存入 httpOnly Cookie | JS 无法读取 Token，**架构级防 XSS** | `c.SetCookie` 参数检查 |
| S06 | Cookie 设置 `SameSite=Lax` | 防 CSRF（阻止跨站 POST 请求携带 Cookie） | 响应头检查 |
| S07 | 生产环境 Cookie `Secure=true` | 仅 HTTPS 传输 Cookie，防中间人窃取 | `Validate()` 强校验 |
| S08 | RT Cookie Path 限制为 `/api/v1/admin/auth` | RT 不随普通业务请求携带，缩小泄露面 | 响应头检查 |
| S09 | 解析时校验签名算法 | 防 `alg: none` 攻击 | `ParseToken` 内 `SigningMethodHMAC` 断言 |
| S10 | Access Token 短有效期（≤ 2h） | 泄露窗口最小化 | 配置 + 测试 |
| S11 | 响应体不返回 Token 和密码哈希 | Token 仅通过 Set-Cookie 传输；`json:"-"` 屏蔽密码 | 接口测试 |
| S12 | 密码明文不写入日志 | `slog` 中避免 log 请求 body | 代码 review |
| S13 | 生产环境 `gin.ReleaseMode` | 防止调试信息泄露 | bootstrap 配置 |
| S14 | 登出接口清除双 Cookie | `MaxAge=0` 确保浏览器立即删除 | 接口测试 |

### CSRF 防护说明

Cookie 方案的核心顾虑是 CSRF（跨站请求伪造）。本方案通过以下策略防御：

**第一道防线：`SameSite=Lax`**

| SameSite 值 | 效果 |
|-------------|------|
| `Strict` | 任何跨站请求都不携带 Cookie（安全但影响顶级导航体验） |
| **`Lax`** | **顶级导航 GET 可携带，跨站 POST/PUT/DELETE 不携带**（推荐） |
| `None` | 全部携带（需 Secure，不推荐） |

管理后台所有写操作（创建、修改、删除、发布）均为 POST/PUT/DELETE，`SameSite=Lax` 下跨站发起的这些请求**不会携带 Cookie**，CSRF 攻击无效。

**第二道防线（V1.5 可选）：CSRF Token**

如果后续安全审计要求更高，可叠加 CSRF Token：
1. 登录成功时，服务端在响应 body 中返回一个 `csrfToken`（非 httpOnly，前端可读）
2. 前端将 `csrfToken` 存入内存，每次写请求在 Header 中携带 `X-CSRF-Token`
3. 服务端中间件校验 Header 中的 CSRF Token 与 Cookie 中的值是否匹配

**V1 结论**：`SameSite=Lax` + `httpOnly` + `Secure`（生产）已满足管理后台的安全要求，无需额外 CSRF Token。

---

## 八、前端（React 管理后台）对接契约

### 8.1 Axios 全局配置

Cookie 方案下前端最大的变化是：**不需要手动管理 Token**。只需确保 Axios 携带 Cookie：

```typescript
// api/client.ts
const apiClient = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,  // 关键：允许跨域请求携带 Cookie
});
```

> `withCredentials: true` 使浏览器在请求时自动附带 Cookie。如果前后端同域（Vite proxy / Nginx 反代），此项可省略，但建议显式声明。

### 8.2 登录流程

```
1. 用户输入用户名密码 → POST /api/v1/admin/auth/login { username, password }
2. 成功 → 服务端通过 Set-Cookie 写入 access_token + refresh_token
   → 前端无需存储任何 Token，浏览器自动管理
   → 只需从 response.data.expiresIn 获取过期时间（可选，用于主动续签倒计时）
3. 后续请求浏览器自动携带 Cookie，前端不需要任何拦截器注入 Header
```

**与 Bearer 方案的对比**：

| 环节 | Bearer（改前） | Cookie（改后） |
|------|---------------|---------------|
| 登录后存储 | `localStorage.setItem('token', ...)` | **无操作**，浏览器自动存 Cookie |
| 请求携带 | Axios 拦截器手动注入 `Authorization` | **无操作**，浏览器自动携带 Cookie |
| Token 续签 | 从 localStorage 读 RT，手动调 refresh | 浏览器自动携带 RT Cookie，调 refresh 即可 |
| 登出 | `localStorage.removeItem('token')` | 调 `POST /auth/logout`，服务端清除 Cookie |

### 8.3 Token 续签（Axios 拦截器）

前端仍需响应拦截器处理 AT 过期的自动续签，但**不需要手动管理 Token 值**：

```typescript
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    const data = error.response?.data;

    // AT 过期，自动用 RT 续签（RT 由浏览器自动携带 Cookie）
    if (data?.code === 'A00001' && data?.message?.includes('过期') && !originalRequest._retry) {
      originalRequest._retry = true;
      try {
        await apiClient.post('/admin/auth/refresh');  // 无需传 body，RT 在 Cookie 中
        return apiClient(originalRequest);             // 重发原始请求
      } catch {
        // RT 也过期，跳转登录页
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);
```

### 8.4 登出

调用服务端登出接口，由服务端清除 Cookie：

```typescript
async function logout() {
  await apiClient.post('/admin/auth/logout');
  window.location.href = '/login';
}
```

> **不要**试图用前端 JS 清除 httpOnly Cookie——JS 根本读不到它们，`document.cookie` 中不可见。登出必须由服务端 `Set-Cookie: Max-Age=0` 完成。

### 8.5 判断登录状态

由于前端 JS 无法读取 httpOnly Cookie，判断登录状态的推荐方式：

| 方式 | 说明 |
|------|------|
| 调用 `GET /auth/me` | 最可靠，成功即已登录，失败即未登录。App 初始化时调一次 |
| 本地标志位 | 登录成功后 `localStorage.setItem('isLoggedIn', 'true')`，仅做 UI 快判，不代表 Token 有效 |

---

## 九、实现顺序建议

按照「先底层后上层、先编译通过后逻辑对接」的原则：

```
第 1 步：安装依赖
  go get github.com/golang-jwt/jwt/v5
  go get golang.org/x/crypto

第 2 步：pkg/jwtutil（纯函数，无外部依赖，可独立编译验证）
  └── Claims 定义、GenerateToken、ParseToken

第 3 步：domain/model/admin_user.go（领域模型）
  └── AdminUser 结构体、TableName()

第 4 步：repository/postgres/admin_user_repo.go（数据访问）
  └── FindByUsername、FindByID

第 5 步：configs 变更（扩展 JWTConfig、Validate、默认值）

第 6 步：service/auth/service.go（业务编排）
  └── Login、RefreshToken、GetCurrentUser

第 7 步：middleware/jwt_auth.go（认证 + 角色中间件）
  └── JWTAuth、RequireRole

第 8 步：handler/auth.go（HTTP 入口）
  └── Login、Refresh、Me

第 9 步：router.go 改造（auth 公开组 + admin 受保护组）

第 10 步：bootstrap.go 改造（传递 cfg 到 Router）

第 11 步：种子数据（初始管理员账号）

第 12 步：自测验收
```

---

## 十、验收检查清单

### 功能验收

| # | 场景 | 预期结果 | ✓ |
|---|------|---------|---|
| 1 | `POST /auth/login` 正确用户名密码 | 返回 `000000`；响应头含两个 `Set-Cookie`（AT + RT） | ☐ |
| 2 | `POST /auth/login` 错误密码 | 返回 `A00001` + "用户名或密码错误"；无 `Set-Cookie` | ☐ |
| 3 | `POST /auth/login` 不存在的用户名 | 返回 `A00001` + "用户名或密码错误"（不泄露具体原因） | ☐ |
| 4 | `GET /admin/vehicles` 无 Cookie | 返回 `A00001` + "缺少有效的认证令牌" | ☐ |
| 5 | `GET /admin/vehicles` 带有效 AT Cookie | 正常返回车型列表 | ☐ |
| 6 | `GET /admin/vehicles` AT Cookie 已过期 | 返回 `A00001` + "认证令牌已过期" | ☐ |
| 7 | `GET /admin/vehicles` AT Cookie 被篡改 | 返回 `A00001` + "认证令牌无效" | ☐ |
| 8 | `POST /auth/refresh` 带有效 RT Cookie | 响应头含新的 AT `Set-Cookie` | ☐ |
| 9 | `POST /auth/refresh` RT Cookie 已过期 | 返回 `A00001`，需重新登录 | ☐ |
| 10 | `GET /auth/me` 带有效 AT Cookie | 返回当前用户信息（不含 passwordHash） | ☐ |
| 11 | `GET /public/vehicles` 无 Cookie | 正常返回（公开接口无需认证） | ☐ |
| 12 | `GET /health` 无 Cookie | 正常返回（健康检查无需认证） | ☐ |
| 13 | `POST /upload/images` 无 Cookie | 返回 `A00001`（上传需要认证） | ☐ |
| 14 | `POST /auth/logout` | 响应头含两个 `Set-Cookie: Max-Age=0`；后续请求失去认证 | ☐ |

### Cookie 属性验收

| # | 检查项 | ✓ |
|---|--------|---|
| 1 | AT Cookie `HttpOnly=true` | ☐ |
| 2 | RT Cookie `HttpOnly=true` | ☐ |
| 3 | AT Cookie `Path=/api` | ☐ |
| 4 | RT Cookie `Path=/api/v1/admin/auth` | ☐ |
| 5 | 两个 Cookie `SameSite=Lax` | ☐ |
| 6 | 生产环境两个 Cookie `Secure=true` | ☐ |
| 7 | 登录响应 body **不包含** Token 值（仅在 Set-Cookie 中） | ☐ |
| 8 | 浏览器 DevTools → Application → Cookies 可见两个 Cookie，JS `document.cookie` 中**不可见** | ☐ |

### 安全验收

| # | 检查项 | ✓ |
|---|--------|---|
| 1 | `.env` 在 `.gitignore` 中 | ☐ |
| 2 | YAML 配置中 `jwt.secret` 为空 | ☐ |
| 3 | 登录响应 body 不包含 `passwordHash` 字段 | ☐ |
| 4 | `/auth/me` 响应不包含 `passwordHash` 字段 | ☐ |
| 5 | 无 JWT Secret 时服务启动报错（fail fast） | ☐ |
| 6 | 日志中无密码明文 | ☐ |
| 7 | 生产环境 `cookie_secure=false` 时启动报错 | ☐ |

---

## 十一、常见踩坑提醒

### 坑 1：`alg: none` 攻击

JWT 规范允许 `alg` 为 `none`（无签名）。如果解析时不校验算法，攻击者可以伪造任意 Token。**必须在 `ParseToken` 中断言 `SigningMethodHMAC`**。

### 坑 2：bcrypt 哈希长度

bcrypt 输出固定 60 字符。数据库 `password_hash` 列类型为 `TEXT`，没有长度限制——这没问题。但如果未来改为 `VARCHAR`，最小长度应为 `72`（bcrypt 输入上限 72 字节）。

### 坑 3：Token 中的 `exp` 时区

`jwt.NewNumericDate(time.Now())` 使用 UTC 时间戳（Unix 秒数），与时区无关。**不要用 `time.Now().Local()`**。

### 坑 4：Gin 路由注册顺序

Gin 的 radix tree 要求**字面量路由在参数路由之前注册**。auth 路由 `/admin/auth/login` 必须在 `/admin/:resource_id` 之前注册，否则会被参数路由吞掉。当前 router 结构不存在此问题（user/vehicles/categories 都是独立子分组），但扩展时注意。

### 坑 5：Cookie 跨域与本地开发

本地开发时前端（Vite `localhost:5173`）和后端（Gin `localhost:8080`）端口不同，属于**同源但不同端口**。Cookie 的处理：

| 场景 | Cookie 是否自动携带 | 解决方案 |
|------|---------------------|---------|
| Vite proxy 反代到后端 | 是（同源） | `vite.config.ts` 中 `proxy: { '/api': 'http://localhost:8080' }`（**推荐**） |
| 前后端直连（不同端口） | 需 `withCredentials: true` + 后端 CORS 配置 | 需额外配置 `Access-Control-Allow-Credentials: true` |
| 生产 Nginx 反代 | 是（同域名） | 无需特殊处理 |

**推荐本地开发方案**：Vite proxy 将 `/api` 代理到 Go 后端，前后端对浏览器来说是同源，Cookie 自动携带，无需任何 CORS 配置。

### 坑 6：`c.SetCookie` 的 SameSite 参数

Gin 的 `c.SetCookie()` 不直接支持 `SameSite` 属性。需要使用标准库 `http.SetCookie()`：

```go
http.SetCookie(c.Writer, &http.Cookie{
    Name:     "access_token",
    Value:    token,
    Path:     "/api",
    MaxAge:   maxAge,
    HttpOnly: true,
    Secure:   cfg.JWT.CookieSecure,
    SameSite: http.SameSiteLaxMode,  // Gin 的 c.SetCookie 没有此参数
})
```

> 这是 Cookie 方案下最常被遗漏的细节。如果不设 `SameSite`，不同浏览器有不同默认值，可能导致 Cookie 被跨站请求携带。

### 坑 7：登出后「幽灵 Cookie」

如果 `Set-Cookie` 清除时的 `Path` 与设置时不一致，浏览器会认为是不同的 Cookie，导致旧 Cookie 残留。**清除 Cookie 时 Path 必须与设置时完全一致**：

```
设置 AT: Path=/api
清除 AT: Path=/api          ✅ 匹配，Cookie 被删除
清除 AT: Path=/              ✗ 不匹配，旧 Cookie 残留
```

---

## 十二、后续演进路线

| 阶段 | 内容 | 触发时机 |
|------|------|---------|
| V1 当前 | JWT 双 Token（httpOnly Cookie）+ bcrypt + 简化 RBAC + SameSite 防 CSRF | 本次实现 |
| V1.5 | Redis Token 黑名单（登出/踢人即时生效，配合 Cookie 清除双保险） | 接入 Redis 后 |
| V1.5 | CSRF Token 叠加（如安全审计要求更高） | 安全审计后 |
| V1.5 | 登录限流（IP + 账号维度，防暴力破解） | 上线前 |
| V1.5 | 审计日志（`audit_logs` 记录登录/登出事件） | 审计需求明确后 |
| V2 | 客户登录（微信 OpenID），与 admin JWT 隔离 | 客户域上线时 |
| V2+ | RSA 非对称签名（公钥分发给网关/微服务验签） | 微服务拆分时 |

---

*创建日期：2026-04-10*
