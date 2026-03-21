# 前端视角看 Go 后端链路

> 适合现在的我反复回看：先不背 Go 术语，先把 `server/` 理解成“路由分发 + 处理函数 + 业务层 + 数据层”。

## 先记住一句话

一次请求在后端里通常这样流动：

```text
浏览器/小程序请求
-> router 找到处理函数
-> handler 接收请求并调用业务
-> service 组织业务逻辑
-> repository 访问数据库
-> model/rule 提供数据结构和规则
-> response 包成统一 JSON 返回给前端
```

如果换成前端语言：

| Go 后端层 | 前端类比 |
|------|------|
| `router` | React Router / Next route map |
| `handler` | controller / action / API route handler |
| `service` | useCase / 业务层 / store action |
| `repository` | API 封装层，只是这里请求的是数据库 |
| `model` | TypeScript 类型，外加少量对象方法 |
| `rule` | 纯业务校验函数 |
| `response` | 统一响应格式工具 |
| `bootstrap` | 应用启动与依赖装配入口 |

---

## 目录职责图

以后迷路时，优先按这个顺序找：

```text
server/
├─ cmd/
│  ├─ api/                    程序入口，启动 HTTP 服务
│  └─ migrate/                数据库迁移命令
├─ configs/                   配置加载
├─ internal/
│  ├─ bootstrap/              组装应用：Gin、DB、middleware、router
│  ├─ transport/http/router/  注册 URL -> handler
│  ├─ transport/http/handler/ 处理 HTTP 请求和响应
│  ├─ service/                业务逻辑层
│  ├─ repository/postgres/    数据访问层
│  ├─ domain/model/           领域模型/数据结构
│  ├─ domain/rule/            业务规则
│  └─ infrastructure/         基础设施初始化，如 Postgres 连接
├─ migrations/                建表 SQL
└─ pkg/
   ├─ response/               统一响应
   └─ logger/                 日志和中间件
```

---

## 看链路时只问 4 个问题

以后看任何后端接口，都按这 4 个问题顺着找：

1. 这个 URL 在哪注册的？
2. 它进了哪个 handler？
3. 这个 handler 调了哪个 service？
4. 这个 service 最后怎么查库/改库？

只要顺着这 4 个问题找，链路基本不会丢。

---

## 用一条真实接口看懂链路

先只看当前最完整的一条：

```text
GET /api/v1/public/vehicles
```

### 第 1 站：程序入口

文件：
- `server/cmd/api/main.go`

关键理解：
- 这里像前端的 `main.tsx` / `index.tsx`
- 负责读取配置、创建应用、启动 HTTP 服务

你可以把它想成：

```ts
const config = loadConfig()
const app = bootstrap(config)
app.listen(port)
```

---

### 第 2 站：bootstrap 负责把应用拼起来

文件：
- `server/internal/bootstrap/bootstrap.go`

它做的事：
- 创建 Gin 实例
- 注册日志、中间件、request id
- 初始化数据库连接
- 注册 router

前端类比：

```ts
const app = createApp()
app.use(loggerMiddleware)
app.use(requestIdMiddleware)
const db = createDb()
registerRoutes(app, db)
```

它不做业务，只负责“装配应用”。

---

### 第 3 站：router 决定 URL 进哪个 handler

文件：
- `server/internal/transport/http/router/router.go`

当前目标接口的关键代码是：

```go
public := v1.Group("/public")
public.GET("/vehicles", handler.NewVehicles(r.db).List)
```

前端翻译：

```ts
router.get("/api/v1/public/vehicles", vehiclesHandler.list)
```

你以后看到一个 URL，第一反应就是先来 `router.go` 搜它。

---

### 第 4 站：handler 是离 HTTP 最近的一层

文件：
- `server/internal/transport/http/handler/vehicles.go`

#### 4.1 `NewVehicles(db)` 在干什么

核心代码：

```go
func NewVehicles(db *gorm.DB) *Vehicles {
    repo := postgres.NewVehicleRepo(db)
    return &Vehicles{DB: db, VehicleService: vehicle.NewService(repo)}
}
```

前端翻译：

```ts
function newVehiclesHandler(db) {
  const repo = new VehicleRepo(db)
  const service = new VehicleService(repo)
  return new VehiclesHandler(service)
}
```

这一步是在组装依赖：

```text
handler
-> service
-> repository
-> db
```

#### 4.2 `List(c)` 到底在做什么

核心代码可以理解成：

```go
func (v *Vehicles) List(c *gin.Context) {
    onlyPublished := strings.Contains(c.Request.URL.Path, "/public/")
    items, err := v.VehicleService.List(c.Request.Context(), onlyPublished)
    if err != nil {
        response.FailBusiness(c, err.Error())
        return
    }
    response.Success(c, gin.H{"items": items})
}
```

前端翻译：

```ts
async function listHandler(req, res) {
  const onlyPublished = req.url.includes("/public/")
  const items = await vehicleService.list({ onlyPublished })
  return success(res, { items })
}
```

handler 的职责非常固定：

1. 读请求信息
2. 调 service
3. 处理错误
4. 返回统一 JSON

一句话记忆：

> handler 只负责“接 HTTP，调业务，回 HTTP”。

---

### 第 5 站：service 负责业务逻辑

文件：
- `server/internal/service/vehicle/service.go`

当前 `List` 很薄：

```go
func (s *Service) List(ctx context.Context, onlyPublished bool) ([]*model.Vehicle, error) {
    return s.vehicles.List(ctx, onlyPublished)
}
```

前端会觉得这像“纯转发”，这感觉是对的。

但 service 保留下来的价值是：
- 以后可以加业务规则
- 可以做数据组合
- 可以加权限控制
- 可以做缓存或埋点

前端类比：

```ts
async function list(params) {
  return repo.list(params)
}
```

现在简单，不代表以后不需要这一层。

---

### 第 6 站：repository 真正查数据库

文件：
- `server/internal/repository/postgres/vehicle_repo.go`

核心代码：

```go
func (v *VehicleRepo) List(ctx context.Context, onlyPublished bool) ([]*model.Vehicle, error) {
    q := v.db.WithContext(ctx).Model(&model.Vehicle{}).Order("sort_order DESC, updated_at DESC")
    if onlyPublished {
        q = q.Where("status = ?", string(enum.VehicleStatusPublished))
    }
    var rows []model.Vehicle
    if err := q.Find(&rows).Error; err != nil {
        return nil, err
    }
    ...
}
```

前端翻译：

```ts
async function listVehicles(onlyPublished) {
  let query = db.table("vehicles").orderBy("sort_order desc, updated_at desc")
  if (onlyPublished) {
    query = query.where({ status: "published" })
  }
  return await query.findMany()
}
```

这层你就当成：

> 后端内部的数据请求层

只不过它请求的不是 HTTP，而是 PostgreSQL。

---

### 第 7 站：model 是数据结构

文件：
- `server/internal/domain/model/vehicle.go`

你可以直接当成 TypeScript 类型来理解：

```ts
type Vehicle = {
  id: string
  category_id: string
  name: string
  cover_media_id: string
  price_mode: "show_price" | "phone_inquiry"
  msrp_price: number
  status: "draft" | "published" | "unpublished" | "deleted"
}
```

区别是 Go 里的 `Vehicle` 不只是字段，还能挂方法，比如：
- `Publish()`
- `Unpublish()`

所以它是：
- 数据结构
- 少量领域行为

---

### 第 8 站：response 统一返回格式

文件：
- `server/pkg/response/response.go`

成功时大致会返回：

```json
{
  "code": "000000",
  "message": "success",
  "data": {
    "items": []
  },
  "request_id": "...",
  "timestamp": "..."
}
```

前端拿接口时，只要记住项目统一认这套壳。

---

## 把 `GET /api/v1/public/vehicles` 浓缩成一段前端伪代码

```ts
router.get("/api/v1/public/vehicles", async (req, res) => {
  const onlyPublished = req.url.includes("/public/")
  const items = await vehicleService.list(onlyPublished)
  return success(res, { items })
})
```

Go 后端只是把这段拆成了：

```text
router -> handler -> service -> repository
```

让每层职责更清楚。

---

## “发布车辆”这条链应该怎么接

当前项目里，“发布车辆”的业务核心已经有了，但 HTTP 入口还需要继续补齐。

### 已经存在的部分

文件：
- `server/internal/service/vehicle/service.go`
- `server/internal/domain/rule/vehicle_publish.go`
- `server/internal/domain/model/vehicle.go`
- `server/internal/repository/postgres/vehicle_repo.go`

也就是说，这些能力已经存在：
- 根据 ID 查车辆
- 判断是否允许发布
- 把车辆状态改成 `published`
- 把修改后的车辆保存回数据库

### 它完整接起来后应该长这样

```text
POST /api/v1/admin/vehicles/:vehicle_id/publish
-> router 注册 publish 路由
-> handler.Publish 取 path/body 参数
-> service.Publish 组织发布流程
-> repo.GetById 查车辆
-> rule.CanPublishVehicle 校验规则
-> model.Vehicle.Publish() 改状态
-> repo.Update() 落库
-> response.Success() 返回结果
```

### 每层应该做什么

#### router

概念上应该新增类似：

```go
vehicles.POST("/:vehicle_id/publish", vehiclesHandler.Publish)
```

前端翻译：

```ts
router.post("/api/v1/admin/vehicles/:vehicle_id/publish", publishHandler)
```

#### handler

职责：
- 取 `vehicle_id`
- 取请求体
- 调 service
- 回成功/失败响应

前端伪代码：

```ts
async function publishHandler(req, res) {
  const id = req.params.vehicle_id
  const body = req.body

  await vehicleService.publish(id, {
    hasCoverImage: body.hasCoverImage,
    hasDetailImages: body.hasDetailImages,
    hasRequiredParams: body.hasRequiredParams,
  })

  return success(res, { id })
}
```

#### service

当前核心逻辑已经在：
- `server/internal/service/vehicle/service.go`

它做的事：

1. 先查车辆
2. 调规则判断能不能发布
3. 能发布就改状态
4. 存回数据库

前端伪代码：

```ts
async function publish(id, req) {
  const vehicle = await repo.getById(id)
  const { ok, errors } = canPublishVehicle(vehicle, req)
  if (!ok) throw new Error(errors.join("\n"))
  vehicle.status = "published"
  await repo.update(vehicle)
}
```

#### rule

当前规则文件：
- `server/internal/domain/rule/vehicle_publish.go`

它负责“是否允许发布”，比如：
- 必须是 `draft`
- 名称不能为空
- 分类不能为空
- 必须有封面图
- 必须有详情图
- 必填参数必须完整

前端类比：

```ts
function canPublishVehicle(vehicle, req) {
  const errors = []
  ...
  return { ok: errors.length === 0, errors }
}
```

#### repository

最终负责：
- `GetById`
- `Update`

也就是查数据库和保存数据库。

---

## 以后自己找链路的顺序

### 如果你拿到一个新接口，比如：

```text
POST /api/v1/admin/vehicles/:vehicle_id/publish
```

你就按下面顺序找：

1. 去 `router.go` 搜 `publish`
2. 找到它绑定的 handler 方法
3. 去 handler 看它调了哪个 service
4. 去 service 看业务流程
5. 去 repo 看它最后查了什么表、改了什么字段

### 你可以固定问自己这几个问题

1. URL 在哪注册？
2. handler 是谁？
3. service 做了哪些业务判断？
4. repo 操作了哪张表？
5. 最后返回给前端的 JSON 长什么样？

---

## 当前最值得记住的结论

1. `GET /api/v1/public/vehicles` 这条链已经是完整可读样板。
2. `router` 不做业务，只做分发。
3. `handler` 不写数据库，只负责 HTTP。
4. `service` 才是业务逻辑中心。
5. `repository` 才是真正访问 PostgreSQL 的地方。
6. “发布车辆”这条业务链核心已在 `service + rule + repo`，后面主要是把 `router + handler` 补齐。

---

## 一句话版心法

以后看后端别一上来就试图同时理解所有文件，只要顺着这一条线看：

```text
URL -> handler -> service -> repository -> DB
```

看懂一条，再看下一条，你就不会乱。
