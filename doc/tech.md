两/三轮车渠道分销电子展厅 V1 技术文档（开发可用） - 更新版
版本：V1.0 MVP 更新日期：2026-03-15 适用：前端、小程序开发、后台开发、运维、测试

---

## 1. 架构设计
1.1 技术栈
- 小程序端: 微信原生小程序 + Vant Weapp
- 管理后台: React + TypeScript + Ant Design 5 + Vite + Axios + TanStack Query
- 后端服务: Go + Gin + GORM
- 数据库: PostgreSQL（业务数据） + Redis（缓存/高频读）
- 媒体存储: OSS/COS + CDN
- 安全: JWT 单 Token + 简化 RBAC（超级管理员/运营编辑）
- 审计: audit_logs 表，记录关键操作

1.2 分层设计
- Handler：处理 HTTP 请求，调用 service，返回统一响应结构
- Service：业务逻辑，模板渲染、状态校验、权限校验
- Repository：数据库操作封装
- Domain/Model：实体定义、状态机、枚举

---

## 2. 数据建模
2.1 主表设计
| 表名 | 用途 | 关键字段 | 说明 |
|------|------|----------|------|
| admin_users | 后台用户 | id, username, password_hash, role, status, last_login_at | MVP 仅两角色 |
| categories | 分类 | id, parent_id, level, name, status, sort_order | 支持一级/二级分类 |
| param_templates | 参数模板 | id, name, category_id, status | 一级分类绑定模板 |
| param_template_items | 参数项 | template_id, field_key, field_name, field_type, unit, required, display, sort_order | V1 支持文本/数值/单选 |
| vehicles | 车型 | id, category_id, name, cover_media_id, price_mode, msrp_price, status, selling_points, sort_order | cover_media_id 存 **media_assets.id**（UUID）；列表/详情可拼 **coverImageUrl** 展示 |
| vehicle_param_values | 车型参数值 | vehicle_id, template_item_id, value_text | 模板驱动前端渲染 |
| media_assets | 媒体资源 | id, storage_key, mime_type, file_size, asset_type, created_at | 上传成功后落库；**不存完整 URL**，展示时由 storage_key + 公开域名拼接 |
| audit_logs | 审计日志 | id, admin_user_id, action, target_type, target_id, timestamp | 发布/下架/删除 |
| system_settings | 系统全局配置 | id, company_name, customer_service_phone, customer_service_wechat, default_price_mode, disclaimer_text, default_share_title, default_share_image, created_at, updated_at | 新增全局配置表，用于维护后台系统设置和全局策略 |

2.2 数据库索引与性能
- vehicles: (status, sort_order, updated_at)
- vehicle_param_values: (vehicle_id, template_item_id)
- categories: (parent_id, sort_order)
- media_assets: UNIQUE(storage_key)
- system_settings: (id)

---

## 3. API 接口
3.1 统一响应结构
```
{
  "code": "000000",
  "message": "success",
  "data": {},
  "request_id": "xxxxxx",
  "timestamp": "2026-03-15T10:20:30+08:00"
}
```

3.2 错误码
| 错误码 | 分类 | 说明 | HTTP |
|--------|------|------|------|
| 000000 | 成功 | 请求成功 | 200 |
| A00001 | 认证 | token 缺失或非法 | 401 |
| A00004 | 授权 | 无接口访问权限 | 403 |
| B00001 | 参数 | 参数校验失败 | 400 |
| C00002 | 业务 | 当前状态不允许操作 | 409 |
| M00001 | 媒体 | 上传签名获取失败 | 500 |

3.3 管理后台接口
| 功能 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 登录 | POST | /api/v1/admin/auth/login | 返回 JWT Token |
| 刷新 token | POST | /api/v1/admin/auth/refresh | 刷新有效期 |
| 获取当前用户 | GET | /api/v1/admin/auth/me | 返回角色、权限列表 |
| 车型分页列表 | GET | /api/v1/admin/vehicles | 支持分类、状态、标签筛选 |
| 创建草稿 | POST | /api/v1/admin/vehicles | 校验必填项、模板绑定 |
| 发布车型 | POST | /api/v1/admin/vehicles/{id}/publish | 发布成功清理 CDN 缓存 |
| 下架车型 | POST | /api/v1/admin/vehicles/{id}/unpublish | 下架后旧链接兜底 |
| 批量上下架 | POST | /api/v1/admin/vehicles/batch-status | 批量操作 |
| 预览车型 | GET | /api/v1/admin/vehicles/{id}/preview | 临时 preview_token |
| 复制车型 | POST | /api/v1/admin/vehicles/{id}/duplicate | 克隆草稿 |
| 图片上传 | POST | /api/v1/admin/upload/images | multipart `file`；成功 `data` 含 **id**（media_assets 主键）、**url**、**storageKey** |
| 分类管理 CRUD | POST/GET/PUT/DELETE | /api/v1/admin/categories | 新增、编辑、删除、排序 |
| 参数模板 CRUD | POST/GET/PUT/DELETE | /api/v1/admin/param-templates | 模板维护及参数项操作 |
| 系统设置获取 | GET | /api/v1/admin/system-settings | 获取全局配置 |
| 系统设置修改 | PUT | /api/v1/admin/system-settings | 更新全局配置 |

3.4 小程序公开接口
| 功能 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 首页聚合 | GET | /api/v1/public/home | Banner、金刚区、运营专区 |
| 分类选车列表 | GET | /api/v1/public/vehicles | category_id/tag 筛选，分页返回 published |
| 分享落地校验 | GET | /api/v1/public/vehicles/{id}/share-check | 下架兜底文案返回 |

---

## 4. 前端设计与交互
4.1 小程序端
- 首页首屏加载 ≤3秒，列表/详情核心内容 ≤2秒
- 触底加载与下拉刷新列表页
- 弱网骨架屏 + 图片占位
- 卖点内容支持 Markdown 或图片数组（需与产品确认富文本替代方案）
- 分享调用小程序原生分享能力
- 微信自定义事件埋点：home_view、banner_click、category_click、zone_click、vehicle_view、phone_click、wechat_click、share_click

4.2 后台管理
- 图片上传前端压缩（>2MB 自动压缩）
- 视频超 20MB 阻断
- 删除、下架操作二次确认
- 角色权限简化为超级管理员/运营编辑
- 系统设置界面支持全局信息修改

---

## 5. 异常与兜底状态
- 分类下无车型 → 空状态提示
- 图片加载失败 → 占位图
- 下架车型访问 → 下架兜底页
- 未配置电话/微信 → 隐藏按钮

---

## 6. 性能与非功能需求
- 首页首屏 ≤3秒，列表/详情 ≤2秒
- 图片压缩 ≤2MB，视频 ≤20MB
- 支持 300 条车型数据长期维护
- 前后台状态同步实时生效
- 前端字体、点击区域、色彩对比符合适老化标准

---

## 7. 迭代扩展设计
- V2：客户登录、千人千面价格、专属报价 → 新建 customers 域
- V3：订货协同、库存状态 → 新建 orders、inventory 域
- V4：营销赋能、数据看板 → 接入分析服务，V1 公共接口不变

---

## 8. 部署与运维
- dev/test/prod 环境隔离
- Docker 镜像化部署，Nginx 反向代理
- 数据库变更通过 Migration 工具管理
- API 成功率、慢 SQL、Redis 异常、上传失败率监控

