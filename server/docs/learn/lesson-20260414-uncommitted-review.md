# 2026-04-14 未提交改动复盘：公开端闭环、详情图与发布校验

> 本文基于当前工作区未提交内容复盘。它不是正式 API 契约源，而是把本轮改动相对上一版审计文档的推进、边界和剩余风险沉淀下来，方便后续联调与验收。

---

## 一、范围边界

本轮未提交改动覆盖 `server` 与 `admin` 两侧，主线是把 2026-04-14 审计中暴露的几个断点往前推进：

- 修复 Go 编译 P0：`RequiredField` 不再保留无参数 `%s` 占位符，并对字符串做 `TrimSpace`。
- 收口参数模板链路：`PUT /admin/param-templates/:id` 以 URL id 为权威 id，列表空结果返回稳定 `{ list, page }`，`ItemsCount` 不再吞错。
- 补公开端入口：新增独立 `Public` handler，注册 `home/categories/vehicles/detail/contact/share-check`。
- 增加车型详情图链路：新增 `vehicle_detail_media` 表、模型、repo/service、管理端 GET/PUT 接口与后台多图上传组件。
- 推进发布校验：发布前真实检查封面图、启用分类、详情图、启用参数模板下的必填参数值。
- 修复分类分页：handler 使用归一化后的 `pageSize`，不再把默认值算出来后又传回原始 0。

本轮仍不等于完整 V1 交付。特别是车型参数值 `vehicle_param_values` 已建表并用于读取/校验，但当前 diff 中还没有看到管理端录入参数值的完整写入接口与页面闭环。

---

## 二、关键改动对照

| 模块 | 本轮变化 | 链路状态 |
|---|---|---|
| 编译阻断 | `helper.RequiredField` 删除错误 `%s` 占位并补齐换行/格式 | P0 已修复，仍需 `go test ./...` 验证 |
| 参数模板 | handler 引入请求 DTO，repo `Update(ctx, id, p)` 使用 URL id，事务内更新主表、子项 upsert、删除移除项 | URL id 契约基本收口 |
| 参数模板列表 | service 层兜底 `page/pageSize`，空数据返回 `list: []` 和分页元信息，`ItemsCount` 串行并传播错误 | 响应契约更稳定 |
| 分类列表 | `PageSize` 使用归一化后的 `pageSize` | 默认分页行为修复 |
| 车型详情图 | 新增 `vehicle_detail_media`，管理端 `GET/PUT /admin/vehicles/:vehicle_id/detail-images`，前端 `MultiImageUploader` 支持最多 9 张、拖拽排序 | 管理端详情图链路成型 |
| 发布校验 | 从 DB 读取详情图、分类状态、必填参数填写情况，不再写死 true | 从 MVP 放宽进入真实校验 |
| 公开端 | 新增 `GET /public/home/categories/vehicles/{id}/contact/share-check` 等路由，独立 `Public` handler 强制 published 过滤 | 小程序核心读取入口补齐 |
| 系统设置展示 | 公开 contact 与默认分享图通过 `media_assets.storage_key` 拼 OSS URL | 展示层 URL 与存储 key 继续分离 |
| 管理后台 | 新增详情图上传、保存与回显；车型新增/编辑后追加保存详情图 | 页面能力补齐，但保存是两次请求 |

---

## 三、接口契约摘要

### 参数模板

```text
GET    /api/v1/admin/param-templates/list?page=1&pageSize=10
POST   /api/v1/admin/param-templates
PUT    /api/v1/admin/param-templates/:id
GET    /api/v1/admin/param-templates/getById/:id
GET    /api/v1/admin/param-templates/getItemsById/:id
GET    /api/v1/admin/param-templates/getItemsbyId/:id
DELETE /api/v1/admin/param-templates/:id
```

学习点：

- `PUT /:id` 的 id 现在应视为唯一权威来源，body 不再需要传 `id`。
- `getItemsById` 是新补的正确大小写别名，`getItemsbyId` 仍保留，像是在兼容旧前端调用。
- 列表 DTO 已从 domain model 中抽离到 service/handler 边界，避免把 HTTP query 语义污染到领域模型。

### 管理端车型详情图

```text
GET /api/v1/admin/vehicles/:vehicle_id/detail-images
PUT /api/v1/admin/vehicles/:vehicle_id/detail-images
```

`PUT` 请求体：

```json
{
  "images": [
    { "mediaId": "uuid" }
  ]
}
```

服务端采用“整组替换”语义：先删除该车型已有详情图，再按请求顺序重新插入，并用 `sort_order = len(images) - index` 保存展示顺序。这个模式适合前端把详情图当作一个数组状态来编辑。

### 公开端

```text
GET /api/v1/public/home
GET /api/v1/public/categories
GET /api/v1/public/vehicles
GET /api/v1/public/vehicles/:id
GET /api/v1/public/vehicles/:id/share-check
GET /api/v1/public/contact
```

公开端的关键约束：

- 车辆列表和详情只返回 `published` 车型。
- 详情页会返回 `vehicle/contact/detailImages/params/emptyParamsText`。
- `share-check` 在车型下架或不存在时返回 `available: false`，并用系统默认分享标题/图片兜底。
- `home` 当前返回 `banners: []`、`zones: []`，说明首页分区和 Banner 仍是占位。

---

## 四、发布校验的新边界

旧逻辑里 `HasDetailImages` 和 `HasRequiredParams` 被写死为 true，本轮改成从数据库读取真实依赖：

```text
Publish
  -> GetById(vehicle)
  -> GetById(category) 并确认分类启用
  -> HasDetailImages(vehicle_id)
  -> RequiredParamsComplete(vehicle_id, category_id)
  -> rule.CanPublishVehicle
  -> Update(status = published)
```

这让发布从“按钮能点”变成“内容质量有最低门槛”。不过要注意一个新边界：`RequiredParamsComplete` 依赖 `vehicle_param_values` 中已有参数值，而本轮还没有完整看到参数值录入链路。也就是说，发布校验已经变严格，但参数录入能力还需要尽快补齐，否则有启用模板和必填项的车型会被正确拦住，却无法在管理端完成补录。

---

## 五、仍未闭环的问题

### 1. 车型参数值写入链路缺失

`vehicle_param_values` 已经建表，公开详情页也会读取展示参数，但当前改动里没有看到：

```text
GET/PUT /api/v1/admin/vehicles/:id/params
```

或同等能力。后续需要补齐管理端按分类模板渲染参数项、保存 `template_item_id -> value_text`、发布前展示缺失项。

### 2. 车型保存与详情图保存不是一个事务

管理端当前流程是先 `POST/PUT /admin/vehicles`，再调用 `PUT /admin/vehicles/:id/detail-images`。这对前端实现简单，但会出现部分成功：

```text
车型保存成功
详情图保存失败
```

短期可以接受，但需要在 UI 上明确失败提示并允许重试；长期可以考虑把详情图放进车型编辑聚合接口，或者提供更明确的保存状态。

### 3. 首页 Banner 和专区仍是空数组

`GET /public/home` 已经补齐聚合入口，但 `banners`、`zones` 仍是 `[]`。这适合作为 V1 占位，但不要误判为首页展示能力已完整落地。

### 4. 公开 contact 对系统设置错误静默降级

`contact()` 对 `gorm.ErrRecordNotFound` 和其他错误都返回空 contact。公开端可用性更好，但排障上可能掩盖真实 DB/配置问题。后续建议至少打日志，或区分“未配置”和“查询失败”。

### 5. 参数模板路由命名仍需统一

`getItemsById` 与 `getItemsbyId` 双路由并存可以降低兼容风险，但最终应统一到一种命名风格。学习文档里要把它标为“兼容期”，避免新代码继续扩散旧拼写。

### 6. 测试门禁仍未形成

本轮修复了之前已知编译阻断点，但真正可提交前仍要执行：

```text
go fmt ./...
go test ./...
```

如果前端改动也纳入本次提交，还应至少跑管理端构建或类型检查，避免 `MultiImageUploader`、接口类型和表单值之间出现 TS 断点。

---

## 六、本轮学习点

- 文档复盘要跟着真实代码状态更新。昨天的“P0 未修复”“公开端缺失”“参数模板 Update 断裂”在本轮已经部分或全部推进，README 不能继续停留在旧结论。
- URL path id 应该是资源更新的权威 id，body id 最多用于只读回显，不能让两者共同决定更新目标。
- 公开端最好用独立 handler/service 聚合，而不是在管理端 handler 里用 path 判断当前是不是 public 请求。
- 多图编辑适合“数组整体替换 + DB 事务”模式；它比逐张增删接口更贴近前端表单状态，也更容易保证顺序一致。
- 发布校验必须读取真实依赖。写死 true 虽然能让流程跑通，但会把脏数据推到公开展厅。
- 建表只是半条链路。`vehicle_param_values` 只有迁移、模型、读取、校验还不够，必须有管理端写入和回显，才算参数能力闭环。

---

## 七、建议下一步

1. 运行 `go fmt ./...` 与 `go test ./...`，确认 P0 编译阻断已真正解除。
2. 补齐车型参数值管理接口和后台表单，让 `vehicle_param_values` 有写入来源。
3. 为 `Publish` 增加至少一组表驱动测试，覆盖无封面、无详情图、分类禁用、必填参数缺失、全部满足。
4. 给公开端 `VehicleDetail`、`ShareCheck` 增加最小集成测试，确认 draft/unpublished/deleted 不会泄漏。
5. 决定 `getItemsById/getItemsbyId` 的兼容窗口，后续只保留一个规范路由。
