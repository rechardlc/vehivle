# pkg/logger

## 作用
提供统一日志门面和格式约定。

## 建议输出字段
- level
- timestamp
- request_id
- user_id
- action
- latency_ms
- error

## 企业实践
- 开发环境可读，生产环境结构化 JSON。
- 避免打印敏感信息（密码、token、手机号明文）。
