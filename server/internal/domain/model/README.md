# internal/domain/model

## 作用
定义领域实体（不等于 ORM 结构体）。

## V1 建议实体
- AdminUser
- Category
- ParamTemplate / ParamTemplateItem
- Vehicle / VehicleParamValue
- MediaAsset
- SystemSetting
- AuditLog

## 设计建议
- 字段命名贴近业务语义。
- 对外暴露的方法表达业务行为，如 `Publish()`、`Unpublish()`。
