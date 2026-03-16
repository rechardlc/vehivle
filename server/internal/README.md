# internal

## 作用
后端核心实现层，对外不可直接复用。

## 分层说明
- `bootstrap`：应用初始化与依赖装配。
- `transport`：HTTP 适配层。
- `service`：业务编排层。
- `repository`：数据访问层。
- `domain`：领域模型、规则和枚举。
- `integration`：外部系统集成。

## 边界纪律
- Handler 不直接操作数据库。
- Service 不依赖具体框架细节。
- Repository 不写业务流程判断。
- Domain 不依赖 Gin/GORM 等外部实现。
