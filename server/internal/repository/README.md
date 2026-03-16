# internal/repository

## 作用
定义并实现数据访问能力。

## 分工
- `postgres`：业务主数据持久化。
- `redis`：缓存与热点读。

## 规范
- Repository 只关注数据读写，不做业务决策。
- 提供事务接口，供 service 在需要时组合多个写操作。
- 所有查询都应可追踪到 PRD 的字段和筛选需求。
