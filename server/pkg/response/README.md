# pkg/response

## 作用
统一 API 返回体，严格对齐 tech 文档。

## 标准结构
- code
- message
- data
- request_id
- timestamp

## 实施建议
- 成功和失败都统一从此目录输出。
- 不允许 handler 手写散乱 JSON 结构。
