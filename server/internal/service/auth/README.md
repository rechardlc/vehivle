# internal/service/auth

## 作用
后台认证与会话管理

## 业务范围
登录、刷新 Token、获取当前用户、密码策略、角色信息装载。

## 关键规则
账号状态校验、JWT 过期处理、登录失败审计。

## 学习建议
先把该域的领域规则写清楚，再实现 repository 接口，最后接 handler。
