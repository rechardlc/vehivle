# Go 学习指南：前端开发者通过 vehivle 项目入门

> 面向有前端背景的开发者，以「自己写代码 + AI 辅助」的方式学习 Go 后端开发。

---

## 一、项目背景与你的优势

### 1.1 项目技术栈（来自 tech.md）

- **后端服务**：Go + Gin + GORM
- **数据库**：PostgreSQL（业务数据）+ Redis（缓存）
- **分层设计**：Handler → Service → Repository → Domain/Model

### 1.2 前端背景带来的优势

- 熟悉 API 调用、请求/响应结构
- 理解前后端协作、状态管理
- 更容易理解「后端在做什么」

---

## 二、AI 的定位：辅助而非代写

### 2.1 建议的 AI 使用方式

| 用途       | 示例提问方式                                                                 |
|------------|-----------------------------------------------------------------------------|
| 解释概念   | 「Go 的 `context.Context` 在 HTTP 里怎么用？和 JS 的 AbortController 有什么类似？」 |
| 代码审查   | 「这段代码有没有问题？有没有更符合 Go 习惯的写法？」                         |
| 调试协助   | 「这个 panic/编译错误是什么意思？可能出在哪儿？」                           |
| 设计建议   | 「按 tech.md 的分层，这个接口应该放在 Handler 还是 Service？」               |
| 文档对照   | 「tech.md 里这个接口的响应结构，我这样实现对吗？」                           |

### 2.2 应避免的用法

- 「帮我写一个完整的登录接口」
- 「生成整个 vehicles 的 CRUD」

---

## 三、分阶段学习路径

### 阶段 0：Go 基础（约 1–2 周）

**目标**：能看懂 Go 代码、会写简单程序。

**建议**：

1. 安装 Go，跑通 `go run`、`go build`
2. 掌握：变量、类型、结构体、接口、错误处理、goroutine 基础
3. 用「前端思维」对照：
   - `struct` ≈ TypeScript 的 interface
   - `error` ≈ try/catch 的返回值形式
   - `goroutine` ≈ 异步/并发

**可向 AI 提问**：

- 「Go 的 `error` 和 JS 的 `throw` 有什么区别？」
- 「`defer` 和 JS 的 `finally` 有什么不同？」

---

### 阶段 1：最小 HTTP 服务（约 1 周）

**目标**：用 Gin 起一个能返回 JSON 的 HTTP 服务。

**建议**：

1. 新建 `server/`，初始化 `go mod init`
2. 引入 Gin，写一个 `GET /api/v1/health` 返回 `{"status":"ok"}`
3. 对照 tech.md 的「统一响应结构」，自己实现一个 `Response` 结构体并封装返回函数

**可向 AI 提问**：

- 「Gin 的 `c.JSON()` 和 `c.ShouldBindJSON()` 分别做什么？」
- 「我想按 tech.md 的格式返回，这样写对吗？」（贴你的代码）

---

### 阶段 2：分层与第一个接口（约 2 周）

**目标**：按 Handler → Service → Repository 实现一个简单接口。

**建议**：

1. 选一个最简单的接口，例如 `GET /api/v1/admin/system-settings`
2. 自己设计目录：`internal/handler`、`internal/service`、`internal/repository`
3. 先写 Model（如 `SystemSettings`），再写 Repository（查库），再写 Service（业务），最后写 Handler（HTTP）

**可向 AI 提问**：

- 「Handler 里应该放业务逻辑吗？还是只做参数校验和调用 Service？」
- 「这段 GORM 查询有没有 N+1 问题？」（贴代码）

---

### 阶段 3：认证与中间件（约 1–2 周）

**目标**：实现 JWT 认证和 RBAC 中间件。

**建议**：

1. 自己实现登录接口（查用户、校验密码、生成 JWT）
2. 写一个 `AuthMiddleware`，从 Header 解析 token 并写入 context
3. 写一个 `RBACMiddleware`，根据角色限制访问

**可向 AI 提问**：

- 「JWT 的 secret 应该放在哪里？环境变量怎么读？」
- 「这段中间件有没有安全漏洞？」（贴代码）

---

### 阶段 4：核心业务接口（约 2–3 周）

**目标**：实现 vehicles、categories、param-templates 等核心接口。

**建议**：

1. 按 tech.md 的表结构建 Model
2. 用 GORM Migration 管理表结构
3. 实现状态机：`draft → published → unpublished → deleted`
4. 自己写参数校验、错误码映射

**可向 AI 提问**：

- 「GORM 的 `Preload` 和 `Joins` 有什么区别？我这个场景用哪个更合适？」
- 「状态机这段逻辑，有没有更清晰的写法？」（贴代码）

---

## 四、具体实践建议

### 4.1 先写再问

- 先自己写一版（哪怕不完整）
- 再问 AI：「这段代码有什么问题？如何改进？」
- 对比 AI 建议，理解差异和原因

### 4.2 用 Cursor 的「解释」功能

- 选中不理解的代码，让 AI 逐行解释
- 问：「这行在做什么？为什么要这样写？」

### 4.3 建立学习笔记

- 在 `doc/` 下建 `go-learning-notes.md`
- 记录：新概念、踩坑、和 JS/TS 的对比
- 定期回顾，形成自己的知识体系

### 4.4 对照接口文档开发

- 用 `doc/接口详细版_请求响应错误码字段校验.docx` 作为契约
- 自己设计请求/响应结构，再和文档对比
- 问 AI：「我这样实现，是否符合文档约定？」

---

## 五、推荐学习资源

| 类型 | 资源 |
|------|------|
| 官方 | [Go 官方教程](https://go.dev/tour/)（A Tour of Go） |
| 书籍 | 《Go 程序设计语言》前几章 |
| Web | [Gin 文档](https://gin-gonic.com/docs/)、[GORM 文档](https://gorm.io/docs/) |
| 视频 | B 站搜索「Go 入门」「Gin 实战」 |

---

## 六、总结

| 维度       | 内容 |
|------------|------|
| 学习顺序   | Go 基础 → Gin HTTP → 分层 → 认证 → 业务接口 |
| AI 用法    | 解释、审查、调试、设计建议，而不是直接生成完整代码 |
| 你的优势   | 熟悉 API、前后端协作，更容易理解 Handler 和接口设计 |
| 项目价值   | 有完整 PRD、tech.md、分层设计，适合边学边做、逐步实现 |

---

## 七、相关文档

- `doc/tech.md`：技术栈、数据建模、API 摘要
- `doc/需求文档_PRD.docx`：产品范围、页面、字段、规则
- `doc/接口详细版_请求响应错误码字段校验.docx`：接口明细、请求/响应示例、错误码
- `.agent/skills/vehivle-v1-showroom/SKILL.md`：项目记忆与开发对齐技能
