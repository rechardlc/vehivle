# internal/integration/object_storage

## 作用
封装对象存储能力。

## V1 需求对应
- 生成前端直传签名（upload-policy）。
- 上传完成后回写 metadata（media/complete）。
- 统一文件类型、大小、扩展名规则。

## 注意事项
- 图片可压缩策略发生在前端，但后端仍需做最终校验。
- 视频超限要明确返回可读错误码。
