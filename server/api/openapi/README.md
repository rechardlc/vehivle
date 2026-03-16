# api/openapi

## 作用
存放 API 规范文件（如 openapi.yaml）。

## V1 建议拆分
- admin-auth
- admin-vehicles
- admin-categories
- admin-param-templates
- admin-media
- admin-system-settings
- public-home-and-vehicles

## 规则
- 文档和实现同步迭代。
- 字段变更必须先更新契约再改代码。
