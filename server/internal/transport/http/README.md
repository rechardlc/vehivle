# internal/transport/http

## 作用
HTTP 接口实现主目录。

## 建议子模块
- `router`：路由注册和分组。
- `middleware`：鉴权、日志、恢复、限流等。
- `handler`：请求处理器。

## 与技术文档对齐
统一返回结构应满足：`code`、`message`、`data`、`request_id`、`timestamp`。
错误码体系应覆盖 A/B/C/M 四大类。
