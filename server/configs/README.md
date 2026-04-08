# configs

## 作用

管理不同环境（dev/prod）的配置加载：YAML 非敏感默认值 + 环境变量敏感信息。

## 加载顺序

1. 加载 `.env`（可选，若存在）
2. 根据 `VEHIVLE_APP_ENV` 或 `APP_ENV` 确定环境（默认 `dev`）
3. 加载 `.env.{env}`（如 `.env.dev` / `.env.prod`）
4. 读取 `conf.{env}.yaml`
5. 环境变量 `VEHIVLE_*` 覆盖同名配置项

## 环境变量命名

| 配置项 | 环境变量 | 示例 |
|--------|----------|------|
| app.env | VEHIVLE_APP_ENV | dev / prod |
| app.port | VEHIVLE_APP_PORT | 9999 |
| database.dsn | VEHIVLE_DATABASE_DSN | postgres://... |
| redis.addr | VEHIVLE_REDIS_ADDR | localhost:6379 |
| redis.password | VEHIVLE_REDIS_PASSWORD | |
| oss.access_key | VEHIVLE_OSS_ACCESS_KEY | |
| oss.secret_key | VEHIVLE_OSS_SECRET_KEY | |
| oss.bucket | VEHIVLE_OSS_BUCKET | vehivle-media |
| oss.region | VEHIVLE_OSS_REGION | ap-guangzhou |
| oss.endpoint | VEHIVLE_OSS_ENDPOINT | http://localhost:9000 |
| oss.public_url | VEHIVLE_OSS_PUBLIC_URL | 浏览器访问 MinIO 的基址；endpoint 为 Docker 内网主机名时必填 |
| oss.enable_public_read | VEHIVLE_OSS_ENABLE_PUBLIC_READ | dev 默认 true（YAML）；直链需匿名读时开启 |
| jwt.secret | VEHIVLE_JWT_SECRET | 32+ 字节密钥 |

## 文件说明

| 文件 | 用途 |
|------|------|
| `conf.dev.yaml` | 开发环境 YAML 默认值 |
| `conf.prod.yaml` | 生产环境 YAML 默认值 |
| `.env.example` | 环境变量模板（复制后填写） |
| `.env.dev` | 开发环境变量（可提交示例值） |
| `.env.prod` | 生产环境变量（敏感值不提交） |

## 实践建议

- 敏感值（DSN、密钥、密码）一律通过环境变量注入，不写死在 YAML。
- 新增配置项时，同步更新 `setDefaults`、YAML 模板和 README。
- 生产部署时设置 `VEHIVLE_APP_ENV=prod` 并加载 `.env.prod`。
