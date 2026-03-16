# internal/integration

## 作用
管理外部系统接入，和内部业务解耦。

## V1 接入对象
- `auth`：JWT、密码哈希。
- `object_storage`：OSS/COS 上传签名与回写。
- `cdn`：发布后缓存刷新。

## 原则
- 对外部 SDK 做二次封装，避免污染 service。
- 接口层返回业务可理解错误，而不是底层报错原文。
