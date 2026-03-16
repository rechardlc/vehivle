# internal/bootstrap

## 作用
负责“把系统拼起来”。

## 这里放什么
- 组件初始化：DB、Redis、Logger、Storage、JWT。
- 依赖注入：将 repository -> service -> handler 串起来。
- 启动健康检查和基础监控。

## 质量要求
- 初始化失败必须快速失败（fail fast）。
- 启动日志应清晰输出当前环境、配置来源、关键依赖状态。
- 所有外部资源连接都要有超时与重试策略。
