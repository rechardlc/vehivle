# internal/transport/http/handler

## 作用
处理 HTTP 入参、调用 service、封装响应。

## Handler 该做的事
- 参数绑定与校验。
- 调用对应 service。
- 把 service 结果映射为统一响应结构。

## Handler 不该做的事
- 不写复杂业务规则。
- 不直连数据库。
- 不处理跨接口的流程编排。

## PRD 覆盖
至少应覆盖 auth、vehicle、category、param template、media、system settings、public home/vehicle/share-check。
