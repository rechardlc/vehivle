# Go 后端学习进度

> 每日进度记录与项目落地状态。学习内容见 [lesson-20260317.md](./lesson-20260317.md)（知识库）、[lesson-20260318.md](./lesson-20260318.md)（今日）。

---

## 一、项目落地顺序（对齐循序渐进总说明）

| 序号 | 内容 | 状态 |
|------|------|------|
| 1 | 第 1 步：工程壳（configs、logger、response、main 启动） | ✅ 已完成 |
| 2 | 第 2 步：HTTP 基座（路由分组、request_id、recovery、access log 中间件） | 进行中 |
| 3 | 第 3 步：业务语义（domain/model、enum、rule） | 待开始 |
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
| 3 | 路由分组（admin/public）、完整中间件链 | 进行中 |
| 4 | PostgreSQL + 迁移 + Repository | 待开始 |
| 5 | 认证与权限（JWT、RBAC） | 待开始 |
| 6 | 核心业务域（分类、车型、媒体等） | 待开始 |
| 7 | 缓存、性能、异常兜底 | 待开始 |
| 8 | 测试与质量门禁 | 待开始 |
| 9 | 部署与上线 | 待开始 |

---

## 三、当前目录结构（带状态）

```
server/
├── go.mod              # 模块定义
├── go.sum              # 依赖校验和
├── cmd/api/main.go     # 入口 ✅（configs.Load、Gin 模式、中间件链、health）
├── configs/            # 配置 ✅（YAML + env、viper、Conf 结构）
├── internal/bootstrap/ # 依赖注入（待落地）
├── pkg/
│   ├── logger/         # 日志 ✅（Logger、RequestID、AccessLog 中间件）
│   └── response/       # 统一响应 ✅
└── ...
```

---

## 四、下一步建议（第 2 步：HTTP 基座）

1. **路由分组**：`/api/v1/admin`、`/api/v1/public`
2. **完善中间件**：request_id、recovery、access log 已就绪，可补充自定义 recovery
3. **bootstrap 抽取**：将 Gin 初始化、路由注册移到 `internal/bootstrap`
4. **端口从配置读取**：`r.Run(":" + strconv.Itoa(cfg.App.Port))`

---

## 五、每日进度

> 按日期倒序，最新在前。建议每天结束前：自测通过、写 5 行复盘、记录明天第一件事。

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

*最后更新：2026-03-17*
