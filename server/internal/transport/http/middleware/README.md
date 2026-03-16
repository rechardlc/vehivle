# internal/transport/http/middleware

## 作用
封装请求链公共能力。

## V1 必备中间件
- RequestID：生成/透传请求 ID。
- Recovery：统一捕获 panic。
- AccessLog：记录访问日志与耗时。
- Auth：JWT 校验。
- RBAC：超级管理员/运营编辑权限控制。

## 注意事项
- 中间件顺序会影响行为，先日志与恢复，再鉴权和权限。
- 错误返回必须走统一错误码和响应结构。
