# JWT 认证闭环实现指导文档

> **对齐**：[循序渐进总说明](./循序渐进总说明.md) **第 5 步——做认证闭环**。  
> **前置**：第 1～4.5 步已完成（工程壳、HTTP 基座、Domain、数据库迁移、对象存储）。  
> **数据库**：`admin_users` 表已在 `000001_init_schema.up.sql` 中建好（UUID 主键、`username`、`password_hash`、`role`）。  
> **配置**：`configs.JWTConfig` 已定义 `Secret` + `ExpireHours`，`.env.example` 预留 `VEHIVLE_JWT_SECRET`。  
> **响应码**：`A00001`（认证失败）、`A00004`（授权失败）及 `FailAuth`/`FailAuthDenied` 已就绪。

| 元数据 | 值 |
|--------|-----|
| **涉及目录** | `pkg/jwt`、`domain/model`、`repository/postgres`、`service/auth`、`transport/http/handler`、`transport/http/middleware`、`bootstrap`、`configs` |
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

### 1.2 单 Token vs 双 Token

技术文档写的是「JWT 单 Token」，但企业实践中推荐 **Access Token + Refresh Token 双 Token 模式**：

| 方案 | Access Token 有效期 | Refresh Token 有效期 | 优势 | 劣势 |
|------|---------------------|----------------------|------|------|
| 单 Token | 较长（24h） | 无 | 简单 | 过期即重新登录；Token 泄露风险窗口大 |
| 双 Token | 较短（2h） | 较长（7d） | AT 泄露窗口小；用户无感续签 | 多一次 refresh 请求 |

**V1 采用双 Token**。即使 V1 先不接 Redis 黑名单，双 Token 的短有效期本身就是一层安全兜底。后续 V1.5 接 Redis 后可叠加「登出吊销」能力。

### 1.3 Token 传输方式：AT 走 Authorization，RT 走 httpOnly Cookie

企业实践中 JWT 的传输方式有两种主流选择：

| 对比项 | Authorization Header（Bearer） | httpOnly Cookie |
|--------|-------------------------------|-----------------|
| XSS 防护 | JS 可读取 Token，需前端避免 XSS 泄露 | httpOnly 标记后 JS 完全无法读取 |
| CSRF 防护 | 天然免疫（不自动携带） | 浏览器自动携带 Cookie，需配合 `SameSite` 防护 |
| 前端复杂度 | 需保存 AT，并由 Axios 拦截器注入 Header | 浏览器自动携带 |
| 登出 | 前端清除 AT，本项目服务端同时清 RT Cookie | 服务端 `Set-Cookie: MaxAge=0` 清除 |
| 移动端兼容 | 原生 App 友好 | WebView 可用，原生 App 需额外处理 |

**本项目当前选择混合方案**：
1. **Access Token（AT）不走 Cookie**：登录/续签响应体返回 `accessToken`，前端缓存后通过 `Authorization: Bearer {accessToken}` 携带。
2. **Refresh Token（RT）走 httpOnly Cookie**：前端 JS 不读取 RT，续签时浏览器自动携带 `refresh_token`。
3. 管理端业务接口统一只认 `Authorization: Bearer ...`，不兼容旧的 `accessToken` Header，也不读取 `access_token` Cookie。
4. RT Cookie 仍配合 `HttpOnly + SameSite=Lax + Secure(生产)`，降低 RT 泄露和 CSRF 风险。

---

## 二、整体流程图

```
┌─────────────┐                           ┌──────────────────────────┐
│  React 管理后台 │                           │      Go Server           │
│              │                           │                          │
│  1. 登录表单  │── POST /auth/login ──────▶│  验证 username+password  │
│              │◀─ Body: accessToken        │  bcrypt 比对              │
│              │   Set-Cookie:             │  签发 AT + RT             │
│              │    refresh_token=xxx;      │  AT 返回 body，RT 写 Cookie│
│              │    HttpOnly; SameSite=Lax  │                          │
│              │                           │                          │
│  2. 请求 API  │── Authorization: Bearer AT▶│  JWT 中间件               │
│              │◀─ 正常业务响应 ────────────│  解析 → 验签 → 注入 Claims │
│              │                           │                          │
│  3. AT 过期   │── POST /auth/refresh ────▶│  从 Cookie 读取 RT       │
│  （浏览器自动  │◀─ Body: 新 accessToken     │  验证 RT 有效性           │
│    携带 RT）  │                           │  签发新 AT → 返回 body    │
│              │                           │                          │
│  4. 获取用户  │── Authorization: Bearer AT▶│  从 Header 读取 AT       │
│  （需 AT）    │◀─ { id, username, role } ─│  解析 Claims → 查库      │
│              │                           │                          │
│  5. 登出     │── POST /auth/logout ─────▶│  Set-Cookie: MaxAge=0    │
│              │◀─ RT Cookie 被清除 ────────│  前端清本地 AT           │
└─────────────┘                           └──────────────────────────┘
```

**关键区别**：AT 出现在登录/续签响应体中，并由前端通过 `Authorization: Bearer ...` 携带；RT 不出现在 JSON 响应体中，只通过 `refresh_token` httpOnly Cookie 传输。

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
│   └── jwt/
│       └── jwt.go                  ← 纯函数：Token 签发、解析、Claims 定义
├── internal/
│   ├── domain/model/
│   │   └── admin_user.go           ← AdminUser 领域模型（对应 admin_users 表）
│   ├── repository/postgres/
│   │   └── admin_user_repo.go      ← 用户数据访问（FindByUsername、FindByID）
│   ├── service/auth/
│   │   └── service.go              ← 认证业务：Login、Refresh、GetCurrentUser
│   └── transport/http/
│       ├── handler/
│       │   └── auth.go             ← HTTP 入口：Login、Refresh、Me、Logout（返回 AT、写/清 RT Cookie）
│       └── middleware/
│           └── jwtAuth.go          ← JWT 认证中间件 + 角色鉴权中间件
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

### 5.1 `pkg/jwt/jwt.go` — JWT 工具包

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
  ├─ 3. jwt.GenerateToken(secret, atDuration, user.ID, user.Username, user.Role)
  │     └─ 签发 Access Token（短有效期，如 2 小时）
  │
  ├─ 4. jwt.GenerateToken(secret, rtDuration, user.ID, user.Username, user.Role)
  │     └─ 签发 Refresh Token（长有效期，如 7 天）
  │
  └─ 5. return { accessToken, refreshToken, expiresIn }
        （Token 字符串返回给 Handler；Handler 将 AT 写入响应体，将 RT 写入 httpOnly Cookie）
```

**密码哈希说明**：

- 使用 `golang.org/x/crypto/bcrypt`，**不要用** MD5/SHA256
- bcrypt 自带盐值，`cost=10`（默认值）在当前硬件下每次哈希约 100ms，足以防暴力破解
- `CompareHashAndPassword` 是**恒定时间比较**，防时序攻击

> 类比前端：bcrypt 类似于前端「不要用 `==` 比较密码」的理念升级版——它不仅比较安全，还自带加盐和慢哈希抗暴力破解。

---

### 5.5 `internal/transport/http/middleware/jwtAuth.go` — 认证中间件

**两个中间件**：

#### （1）`JWTAuth(secret string) gin.HandlerFunc` — JWT 认证

Access Token 从标准 `Authorization` Header 中读取，格式固定为 `Bearer {accessToken}`。不再兼容旧的 `accessToken` Header，也不再读取 `access_token` Cookie。

执行流程：

```
请求进入
  │
  ├─ 1. 从 Header 读取 Authorization
  │     c.GetHeader("Authorization")
  │     └─ 缺失、为空或没有 Bearer 前缀 → FailAuth("缺少有效的认证令牌") + Abort
  │
  ├─ 2. 去掉 "Bearer " 前缀，得到 tokenString
  │
  ├─ 3. jwt.ParseToken(secret, tokenString)
  │     ├─ ErrTokenExpired → FailAuth("认证令牌已过期") + Abort
  │     └─ ErrTokenInvalid → FailAuth("认证令牌无效") + Abort
  │
  ├─ 4. 将 Claims 注入 gin.Context（供后续 Handler 读取）
  │     ├─ c.Set("current_user_id", claims.UserID)
  │     ├─ c.Set("current_username", claims.Username)
  │     └─ c.Set("current_role", claims.Role)
  │
  └─ 5. c.Next()
```

**Header/Context Key 建议定义为常量**：

```
HeaderAuthorization = "Authorization"
BearerPrefix        = "Bearer "
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

| 接口 | 方法 | 路径 | 是否需要 JWT | Token 行为 |
|------|------|------|-------------|------------|
| 登录 | POST | `/api/v1/admin/auth/login` | 否（公开） | 响应体返回 AT；写入 RT Cookie |
| 刷新 | POST | `/api/v1/admin/auth/refresh` | 否（浏览器自动携带 RT Cookie） | 响应体返回新 AT |
| 当前用户 | GET | `/api/v1/admin/auth/me` | 是（`Authorization: Bearer AT`） | 无 |
| 登出 | POST | `/api/v1/admin/auth/logout` | 否 | 清除 RT Cookie；前端清本地 AT |

**请求/响应契约**：

```
# POST /api/v1/admin/auth/login
Request Body:
  { "username": "admin", "password": "123456" }

Response Headers (成功):
  Set-Cookie: refresh_token=eyJhbGci...; Path=/api/v1/admin/auth; HttpOnly; SameSite=Lax; Max-Age=604800

Response Body (成功):
  {
    "code": "000000",
    "message": "success",
    "data": {
      "expiresIn": 7200,
      "accessToken": "eyJhbGci..."
    }
  }

Response Body (失败):
  {
    "code": "A00001",
    "message": "用户名或密码错误"
  }
```

> **注意**：Access Token 必须在 JSON body 中返回，供前端保存并写入 `Authorization: Bearer ...`。Refresh Token 不返回 body，仅通过 httpOnly Cookie 写入。

```
# POST /api/v1/admin/auth/refresh
Request: 无 body（浏览器自动携带 refresh_token Cookie）

Response Headers (成功):
  不设置 Access Token Cookie

Response Body (成功):
  {
    "code": "000000",
    "data": {
      "expiresIn": 7200,
      "accessToken": "eyJhbGci..."
    }
  }
```

> **Refresh Token 的 Cookie Path 限制**：RT Cookie 的 `Path` 设为 `/api/v1/admin/auth`，使其**仅在 auth 路径下被浏览器携带**，不会随普通业务请求发送，缩小泄露面。

```
# GET /api/v1/admin/auth/me
Authorization: Bearer eyJhbGci...

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
  Set-Cookie: refresh_token=; Path=/api/v1/admin/auth; HttpOnly; SameSite=Lax; Max-Age=0

Response Body:
  {
    "code": "000000",
    "message": "success"
  }
```

**Cookie 写入工具函数建议**：

Handler 中只需要管理 Refresh Token Cookie，建议抽取为辅助函数统一管理 Cookie 属性：

```
setRefreshTokenCookie(c, token, maxAge)  → http.SetCookie(...Name: "refresh_token", Path: "/api/v1/admin/auth", HttpOnly: true, SameSite: Lax...)
clearAuthCookies(c)                      → RT Cookie 设 MaxAge=0
```

**`http.Cookie` 关键字段含义**：

| 参数 | 值 | 说明 |
|------|-----|------|
| name | `"refresh_token"` | Cookie 名称 |
| value | Token 字符串 / `""` | Cookie 值（清除时为空串） |
| maxAge | 秒数 / `0` | 有效期（0 = 会话级，-1 或 0 用于删除） |
| path | `"/api/v1/admin/auth"` | Cookie 生效路径范围 |
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
  │     ├── POST /login           返回 AT，写入 RT Cookie
  │     ├── POST /refresh         从 RT Cookie 读取，返回新 AT
  │     └── POST /logout          清除 RT Cookie
  │
  ├── /admin                    ← 受保护（挂载 JWTAuth 中间件，从 Authorization 读取 AT）
  │     ├── GET  /auth/me
  │     ├── /vehicles/...
  │     ├── /categories/...
  │     └── POST /upload/images
  │
  └── /public                   ← 公开（保持不变）
        └── GET /vehicles
```

**Router 结构体需要持有 JWT 配置**：中间件需要 `JWT.Secret`，Auth Handler 需要 Cookie 配置写入/清除 RT Cookie。

**logout 路由**：`POST /admin/auth/logout` 放在 auth 公开组（不依赖 JWT 中间件，服务端只需清除 RT Cookie）。

**依赖注入链路变更**：

```
Bootstrap
  └── injectRouter(gin, logger, db, oss, cfg)   ← 新增 cfg
        └── Router.Register()
              ├── NewAdminUserRepo(db)
              ├── auth.NewService(repo, cfg.JWT)
              ├── handler.NewAuth(authService, cfg.JWT)  ← Handler 需要 JWT 配置以写 RT Cookie
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
    CookieDomain       string `mapstructure:"cookie_domain"`         // RT Cookie 域名（本地留空，生产填域名）
    CookieSecure       bool   `mapstructure:"cookie_secure"`         // RT Cookie 仅 HTTPS 传输（生产 true）
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
// 生产环境强制 RT Cookie Secure
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

#### （5）Token 传输速查表

| 项 | Access Token | Refresh Token | 说明 |
|------|--------------|---------------|------|
| 返回位置 | 登录/续签响应体 `data.accessToken` | `Set-Cookie: refresh_token=...` | AT 给前端注入 Header；RT 不暴露给 JS |
| 请求携带 | `Authorization: Bearer {accessToken}` | 仅 `/api/v1/admin/auth/*` 下自动携带 Cookie | 业务接口不读取 RT |
| 存储位置 | 前端本地缓存 | httpOnly Cookie | 当前 admin 使用 localStorage 保存 AT |
| Cookie Name | 无 | `refresh_token` | 不再设置 `access_token` Cookie |
| Cookie Path | 无 | `/api/v1/admin/auth` | RT 限制在 auth 路径，缩小携带范围 |
| HttpOnly | 无 | `true` | JS 不可读取 RT |
| Secure | 无 | `cfg.JWT.CookieSecure` | 生产 true（仅 HTTPS） |
| SameSite | 无 | `Lax` | 减少跨站请求自动携带 RT 的风险 |
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
| S05 | AT 仅通过 `Authorization: Bearer` 携带 | 避免浏览器对业务请求自动携带 AT，减少 CSRF 面 | 中间件检查 |
| S06 | RT 存入 httpOnly Cookie | JS 无法读取 RT，降低长期凭证泄露风险 | `http.SetCookie` 参数检查 |
| S07 | 生产环境 RT Cookie `Secure=true` | 仅 HTTPS 传输 Cookie，防中间人窃取 | `Validate()` 强校验 |
| S08 | RT Cookie Path 限制为 `/api/v1/admin/auth` | RT 不随普通业务请求携带，缩小泄露面 | 响应头检查 |
| S09 | 解析时校验签名算法 | 防 `alg: none` 攻击 | `ParseToken` 内 `SigningMethodHMAC` 断言 |
| S10 | Access Token 短有效期（≤ 2h） | 泄露窗口最小化 | 配置 + 测试 |
| S11 | 登录/续签响应只返回 AT，不返回 RT 和密码哈希 | RT 保持 httpOnly；`json:"-"` 屏蔽密码 | 接口测试 |
| S12 | 密码明文不写入日志 | `slog` 中避免 log 请求 body | 代码 review |
| S13 | 生产环境 `gin.ReleaseMode` | 防止调试信息泄露 | bootstrap 配置 |
| S14 | 登出接口清除 RT Cookie，前端清 AT | `MaxAge=0` 确保浏览器立即删除 RT | 接口测试 |

### CSRF 防护说明

混合方案下，业务接口的 AT 通过 `Authorization` Header 手动携带，浏览器不会在跨站请求中自动附带，因此天然降低 CSRF 风险。RT Cookie 仍会被浏览器管理，所以需要以下策略：

**第一道防线：`SameSite=Lax`**

| SameSite 值 | 效果 |
|-------------|------|
| `Strict` | 任何跨站请求都不携带 Cookie（安全但影响顶级导航体验） |
| **`Lax`** | **顶级导航 GET 可携带，跨站 POST/PUT/DELETE 不携带**（推荐） |
| `None` | 全部携带（需 Secure，不推荐） |

`refresh_token` Cookie 只用于 `/api/v1/admin/auth/refresh` 等 auth 路径，且 `SameSite=Lax` 下跨站发起的 POST 请求不会携带该 Cookie。普通业务写操作不依赖 Cookie，而是依赖 `Authorization: Bearer ...`。

**第二道防线（V1.5 可选）：CSRF Token**

如果后续安全审计要求更高，可叠加 CSRF Token：
1. 登录成功时，服务端在响应 body 中返回一个 `csrfToken`（非 httpOnly，前端可读）
2. 前端将 `csrfToken` 存入内存，每次写请求在 Header 中携带 `X-CSRF-Token`
3. 服务端中间件校验 Header 中的 CSRF Token 与 Cookie 中的值是否匹配

**V1 结论**：业务接口使用 `Authorization: Bearer`，RT Cookie 使用 `SameSite=Lax` + `httpOnly` + `Secure`（生产），已满足管理后台的基础安全要求，无需额外 CSRF Token。

---

## 八、前端（React 管理后台）对接契约

### 8.1 Axios 全局配置

前端需要同时做两件事：请求时注入 `Authorization: Bearer ...`，并保留 `withCredentials: true` 让浏览器在续签/登出时携带 RT Cookie。

```typescript
// api/client.ts
const apiClient = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,  // 允许 auth/refresh 携带 refresh_token Cookie
});

apiClient.interceptors.request.use((config) => {
  const accessToken = localStorage.getItem('vehivle_admin_access_token');
  if (accessToken) {
    config.headers.set('Authorization', `Bearer ${accessToken}`);
  }
  return config;
});
```

> 业务接口只认 `Authorization: Bearer ...`。不要再发送 `accessToken` 自定义 Header，也不要依赖 `access_token` Cookie。

### 8.2 登录流程

```
1. 用户输入用户名密码 → POST /api/v1/admin/auth/login { username, password }
2. 成功 → 服务端响应 body 返回 accessToken，并通过 Set-Cookie 写入 refresh_token
3. 前端保存 accessToken（当前 admin 使用 localStorage）
4. 后续请求由 Axios 拦截器注入 Authorization: Bearer {accessToken}
```

**当前方案关键行为**：

| 环节 | 行为 |
|------|------|
| 登录后存储 | 前端保存 AT，浏览器保存 RT Cookie |
| 请求携带 | Axios 注入 `Authorization: Bearer AT` |
| Token 续签 | 浏览器携带 RT Cookie，响应体返回新 AT，前端更新本地 AT |
| 登出 | 前端清 AT，服务端清 RT Cookie |

### 8.3 Token 续签（Axios 拦截器）

前端响应拦截器处理 AT 过期的自动续签：调用 refresh 时浏览器自动携带 RT Cookie，成功后将新的 AT 写入本地缓存，再重发原请求。

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
        const refreshResponse = await apiClient.post('/admin/auth/refresh');
        const nextAccessToken = refreshResponse.data.data?.accessToken;
        localStorage.setItem('vehivle_admin_access_token', nextAccessToken);
        return apiClient(originalRequest);
      } catch {
        // RT 也过期，跳转登录页
        localStorage.removeItem('vehivle_admin_access_token');
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);
```

### 8.4 登出

调用服务端登出接口，由服务端清除 RT Cookie，前端同时清除本地 AT：

```typescript
async function logout() {
  await apiClient.post('/admin/auth/logout');
  localStorage.removeItem('vehivle_admin_access_token');
  window.location.href = '/login';
}
```

> 前端只能清除自己保存的 AT；RT 是 httpOnly Cookie，必须由服务端 `Set-Cookie: Max-Age=0` 清除。

### 8.5 判断登录状态

判断登录状态的推荐方式：

| 方式 | 说明 |
|------|------|
| 调用 `GET /auth/me` | 最可靠，成功即 AT 有效，失败则尝试 refresh 或回登录页。App 初始化时调一次 |
| 本地缓存用户信息 | 仅做 UI 快判和减少刷新闪烁，不代表 Token 有效 |

---

## 九、实现顺序建议

按照「先底层后上层、先编译通过后逻辑对接」的原则：

```
第 1 步：安装依赖
  go get github.com/golang-jwt/jwt/v5
  go get golang.org/x/crypto

第 2 步：pkg/jwt（纯函数，无外部依赖，可独立编译验证）
  └── Claims 定义、GenerateToken、ParseToken

第 3 步：domain/model/admin_user.go（领域模型）
  └── AdminUser 结构体、TableName()

第 4 步：repository/postgres/admin_user_repo.go（数据访问）
  └── FindByUsername、FindByID

第 5 步：configs 变更（扩展 JWTConfig、Validate、默认值）

第 6 步：service/auth/service.go（业务编排）
  └── Login、RefreshToken、GetCurrentUser

第 7 步：middleware/jwtAuth.go（认证 + 角色中间件）
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
| 1 | `POST /auth/login` 正确用户名密码 | 返回 `000000`；body 含 `accessToken/expiresIn`；响应头含 RT `Set-Cookie` | ☐ |
| 2 | `POST /auth/login` 错误密码 | 返回 `A00001` + "用户名或密码错误"；无 RT `Set-Cookie` | ☐ |
| 3 | `POST /auth/login` 不存在的用户名 | 返回 `A00001` + "用户名或密码错误"（不泄露具体原因） | ☐ |
| 4 | `GET /admin/vehicles` 无 `Authorization` Header | 返回 `A00001` + "缺少有效的认证令牌" | ☐ |
| 5 | `GET /admin/vehicles` 带有效 `Authorization: Bearer AT` | 正常返回车型列表 | ☐ |
| 6 | `GET /admin/vehicles` AT 已过期 | 返回 `A00001` + "认证令牌已过期" | ☐ |
| 7 | `GET /admin/vehicles` AT 被篡改 | 返回 `A00001` + "认证令牌无效" | ☐ |
| 8 | `POST /auth/refresh` 带有效 RT Cookie | body 含新的 `accessToken/expiresIn`，不设置 AT Cookie | ☐ |
| 9 | `POST /auth/refresh` RT Cookie 已过期 | 返回 `A00001`，需重新登录 | ☐ |
| 10 | `GET /auth/me` 带有效 `Authorization: Bearer AT` | 返回当前用户信息（不含 passwordHash） | ☐ |
| 11 | `GET /public/vehicles` 无 `Authorization` | 正常返回（公开接口无需认证） | ☐ |
| 12 | `GET /health` 无 `Authorization` | 正常返回（健康检查无需认证） | ☐ |
| 13 | `POST /upload/images` 无 `Authorization` | 返回 `A00001`（上传需要认证） | ☐ |
| 14 | `POST /auth/logout` | 响应头含 RT `Set-Cookie: Max-Age=0`；前端清 AT 后后续请求失去认证 | ☐ |

### Header / Cookie 属性验收

| # | 检查项 | ✓ |
|---|--------|---|
| 1 | 业务接口只接受 `Authorization: Bearer {accessToken}` | ☐ |
| 2 | 业务接口不接受 `accessToken` 自定义 Header | ☐ |
| 3 | 服务端不设置 `access_token` Cookie | ☐ |
| 4 | RT Cookie `HttpOnly=true` | ☐ |
| 5 | RT Cookie `Path=/api/v1/admin/auth` | ☐ |
| 6 | RT Cookie `SameSite=Lax` | ☐ |
| 7 | 生产环境 RT Cookie `Secure=true` | ☐ |
| 8 | 登录/续签响应 body 包含 `accessToken`，不包含 `refreshToken` | ☐ |

### 安全验收

| # | 检查项 | ✓ |
|---|--------|---|
| 1 | `.env` 在 `.gitignore` 中 | ☐ |
| 2 | YAML 配置中 `jwt.secret` 为空 | ☐ |
| 3 | 登录响应 body 不包含 `passwordHash` 字段 | ☐ |
| 4 | `/auth/me` 响应不包含 `passwordHash` 字段 | ☐ |
| 5 | 无 JWT Secret 时服务启动报错（fail fast） | ☐ |
| 6 | 日志中无密码明文 | ☐ |
| 7 | 生产环境 `cookie_secure=false` 时启动报错（保护 RT Cookie） | ☐ |

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

### 坑 5：Authorization Header 与 Cookie 跨域配置

本地开发时前端（Vite `localhost:5173`）和后端（Gin `localhost:8080`）端口不同，属于**同源但不同端口**。当前混合方案同时涉及 `Authorization` Header 和 RT Cookie：

| 场景 | 处理方式 |
|------|---------|
| Vite proxy 反代到后端 | 浏览器视角同源，`Authorization` 正常发送，RT Cookie 正常携带（推荐） |
| 前后端直连（不同端口） | 后端 CORS 需允许 `Authorization` Header；如需携带 RT Cookie，还需 `withCredentials: true` + `Access-Control-Allow-Credentials: true` |
| 生产 Nginx 反代 | 同域名部署，通常无需特殊处理 |

**推荐本地开发方案**：Vite proxy 将 `/api` 代理到 Go 后端，前后端对浏览器来说是同源，可减少 CORS 与 Cookie 配置干扰。

### 坑 6：`http.SetCookie` 的 SameSite 参数

Gin 的 `c.SetCookie()` 不直接支持 `SameSite` 属性。RT Cookie 需要使用标准库 `http.SetCookie()`：

```go
http.SetCookie(c.Writer, &http.Cookie{
    Name:     "refresh_token",
    Value:    refreshToken,
    Path:     "/api/v1/admin/auth",
    MaxAge:   maxAge,
    HttpOnly: true,
    Secure:   cfg.JWT.CookieSecure,
    SameSite: http.SameSiteLaxMode,  // Gin 的 c.SetCookie 没有此参数
})
```

> 这是 RT Cookie 下最常被遗漏的细节。如果不设 `SameSite`，不同浏览器有不同默认值，可能导致 Cookie 被跨站请求携带。

### 坑 7：登出后「幽灵 Cookie」

如果 `Set-Cookie` 清除时的 `Path` 与设置时不一致，浏览器会认为是不同的 Cookie，导致旧 Cookie 残留。**清除 Cookie 时 Path 必须与设置时完全一致**：

```
设置 RT: Path=/api/v1/admin/auth
清除 RT: Path=/api/v1/admin/auth  ✅ 匹配，Cookie 被删除
清除 RT: Path=/api                ✗ 不匹配，旧 Cookie 残留
```

---

## 十二、后续演进路线

| 阶段 | 内容 | 触发时机 |
|------|------|---------|
| V1 当前 | JWT 双 Token（AT: Authorization Bearer；RT: httpOnly Cookie）+ bcrypt + 简化 RBAC | 本次实现 |
| V1.5 | Redis Token 黑名单（登出/踢人即时生效，配合 RT Cookie 清除双保险） | 接入 Redis 后 |
| V1.5 | CSRF Token 叠加（如安全审计要求更高） | 安全审计后 |
| V1.5 | 登录限流（IP + 账号维度，防暴力破解） | 上线前 |
| V1.5 | 审计日志（`audit_logs` 记录登录/登出事件） | 审计需求明确后 |
| V2 | 客户登录（微信 OpenID），与 admin JWT 隔离 | 客户域上线时 |
| V2+ | RSA 非对称签名（公钥分发给网关/微服务验签） | 微服务拆分时 |

---

*创建日期：2026-04-10*
